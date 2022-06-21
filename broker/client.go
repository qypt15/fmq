package broker

import (
	"bytes"
	"context"
	"github.com/eapache/queue"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/qypt15/fmq/broker/lib/sessions"
	"github.com/qypt15/fmq/broker/lib/topics"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
	"math/rand"
	"net"
	"regexp"
	"sync"
	"time"
	"unicode/utf8"
	"reflect"
)

const (
	BrokerInfoTopic = "broker000100101info"
	CLIENT          = 0
	ROUTER          = 1
	REMOTE          = 2
	CLUSTER         = 3
)

const (
	_GroupTopicRegexp = `^\$share/([0-9a-zA-Z_-]+)/(.*)$`
)
const (
	Connected    = 1
	Disconnected = 2
)

const (
	awaitRelTimeout int64 = 20
	retryInterval   int64 = 20
)

var (
	groupCompile = regexp.MustCompile(_GroupTopicRegexp)
)

type client struct {
	typ            int
	mu             sync.Mutex
	broker         *Broker
	conn           net.Conn
	info           info
	route          route
	status         int
	ctx            context.Context
	cancelFunc     context.CancelFunc
	session        *sessions.Session
	subMap         map[string]*subscription
	subMapMu       sync.RWMutex
	topicsMgr      *topics.Manager
	subs           []interface{}
	qoss           []byte
	rmsgs          []*packets.PublishPacket
	routeSubMap    map[string]uint64
	routeSubMapMu  sync.Mutex
	awaitingRel    map[uint16]int64
	maxAwaitingRel int
	inflight       map[uint16]*inflightElem
	inflightMu     sync.RWMutex
	mqueue         *queue.Queue
	retryTimer     *time.Timer
	retryTimerLock sync.Mutex
}

type InflightStatus uint8

const (
	Publish InflightStatus = 0
	Pubrel  InflightStatus = 1
)

type inflightElem struct {
	status    InflightStatus
	packet    *packets.PublishPacket
	timestamp int64
}

type subscription struct {
	client    *client
	topic     string
	qos       byte
	share     bool
	groupName string
}

type info struct {
	clientID  string
	username  string
	password  []byte
	keepalive uint16
	willMsg   *packets.PublishPacket
	localIP   string
	remoteIP  string
}

type route struct {
	remoteID  string
	remoteUrl string
}

var (
	DisconnectedPacket = packets.NewControlPacket(packets.Disconnect).(*packets.DisconnectPacket)
	r                  = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func (c *client) init() {
	c.status = Connected
	c.info.localIP, _, _ = net.SplitHostPort(c.conn.LocalAddr().String())
	remoteAddr := c.conn.RemoteAddr()
	remoteNetwork := remoteAddr.Network()
	c.info.remoteIP = ""
	if remoteNetwork != "websocket" {
		c.info.remoteIP, _, _ = net.SplitHostPort(remoteAddr.String())

	} else {
		ws := c.conn.(*websocket.Conn)
		c.info.remoteIP, _, _ = net.SplitHostPort(ws.Request().RemoteAddr)

	}
	c.ctx, c.cancelFunc = context.WithCancel(context.Background())
	c.subMap = make(map[string]*subscription)
	c.topicsMgr = c.broker.topicsMgr
	c.routeSubMap = make(map[string]uint64)
	c.awaitingRel = make(map[uint16]int64)
	c.inflight = make(map[uint16]*inflightElem)
	c.mqueue = queue.New()
}

func (c *client) readLoop() {
	nc := c.conn
	b := c.broker
	if nc == nil || b == nil {
		return
	}
	keepAlive := time.Second * time.Duration(c.info.keepalive)
	timeOut := keepAlive + (keepAlive / 2)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if keepAlive > 0 {
				if err := nc.SetReadDeadline(time.Now().Add(timeOut)); err != nil {
					log.Error("set read timeout error: ", zap.Error(err), zap.String("ClientID", c.info.clientID))
					msg := &Message{
						client: c,
						packet: DisconnectedPacket,
					}
					b.SubmitWork(c.info.clientID, msg)
					return
				}
			}
			packet, err := packets.ReadPacket(nc)
			if err != nil {
				log.Error("read packet error: ", zap.Error(err), zap.String("ClientID", c.info.clientID))
				msg := &Message{
					client: c,
					packet: DisconnectedPacket,
				}
				b.SubmitWork(c.info.clientID, msg)
				return
			}

			if _, isDisconnect := packet.(*packets.DisconnectPacket); isDisconnect {
				c.info.willMsg = nil
				c.cancelFunc()
			}

			msg := &Message{
				client: c,
				packet: packet,
			}
			b.SubmitWork(c.info.clientID, msg)
		}
	}
}

// extractPacketFields function reads a control packet and extracts only the fields
// that needs to pass on UTF-8 validation
func extractPacketFields(msgPacket packets.ControlPacket) []string {
	var fields []string

	// Get packet type
	switch msgPacket.(type) {
	case *packets.ConnackPacket:
	case *packets.ConnectPacket:
	case *packets.PublishPacket:
		packet := msgPacket.(*packets.PublishPacket)
		fields = append(fields, packet.TopicName)
		break

	case *packets.SubscribePacket:
	case *packets.SubackPacket:
	case *packets.UnsubscribePacket:
		packet := msgPacket.(*packets.UnsubscribePacket)
		fields = append(fields, packet.Topics...)
		break
	}

	return fields
}

// validatePacketFields function checks if any of control packets fields has ill-formed
// UTF-8 string
func validatePacketFields(msgPacket packets.ControlPacket) (validFields bool) {

	// Extract just fields that needs validation
	fields := extractPacketFields(msgPacket)

	for _, field := range fields {

		// Perform the basic UTF-8 validation
		if !utf8.ValidString(field) {
			validFields = false
			return
		}

		// A UTF-8 encoded string MUST NOT include an encoding of the null
		// character U+0000
		// If a receiver (Server or Client) receives a Control Packet containing U+0000
		// it MUST close the Network Connection
		// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.pdf page 14
		if bytes.ContainsAny([]byte(field), "\u0000") {
			validFields = false
			return
		}
	}

	// All fields have been validated successfully
	validFields = true

	return
}

func ProcessMessage(msg *Message) {
	c := msg.client
	ca := msg.packet
	if ca == nil {
		return
	}
	if c.typ == CLIENT {
		log.Debug("Recv message:", zap.String("message type", reflect.TypeOf(msg.packet).String()[9:]), zap.String("ClientID", c.info.clientID))
	}

	if !validatePacketFields(ca) {
		_ = c.conn.Close()
		log.Error("Client disconnected due to malformed packet", zap.String("ClientID", c.info.clientID))

		return
	}

	switch ca.(type) {
	case *packets.ConnackPacket:
	case *packets.ConnectPacket:
	case *packets.PublishPacket:
		packet := ca.(*packets.PublishPacket)
		c.ProcessPublish(packet)

	case *packets.PubackPacket:
		packet := ca.(*packets.PubackPacket)
		c.inflightMu.Lock()
		if _, found := c.inflight[packet.MessageID]; found {
			delete(c.inflight, packet.MessageID)
		} else {
			log.Error("Duplicated PUBACK PacketId", zap.Uint16("MessageID", packet.MessageID))
		}
		c.inflightMu.Unlock()
	}
}

func (c *client) ProcessPublish(packet *packets.PublishPacket) {
	switch c.typ {
	case CLIENT:
		c.processClientPublish(packet)
	case ROUTER:
		c.processRouterPublish(packet)
	case CLUSTER:
		c.processRemotePublish(packet)
	}

}

func (c *client) processClientPublish(packet *packets.PublishPacket) {
	topics := packet.TopicName
}
