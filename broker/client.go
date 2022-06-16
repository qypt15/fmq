package broker

import (
	"context"
	"github.com/eapache/queue"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/qypt15/fmq/broker/lib/sessions"
	"github.com/qypt15/fmq/broker/lib/topics"
	"math/rand"
	"net"
	"regexp"
	"sync"
	"time"
)

const (
	BrokerInfoTopic  = "broker000100101info"
	CLIENT				= 0
	ROUTER  			= 1
	REMOTE   			= 2
	CLUSTER				= 3
)

const (
	_GroupTopicRegexp = `^\$share/([0-9a-zA-Z_-]+)/(.*)$`
)
const (
	Connected = 1
	Disconnected = 2
)

const (
	awaitRelTimeout int64  = 20
	retryInterval	int64 = 20
)

var (
	groupCompile = regexp.MustCompile(_GroupTopicRegexp)
)

type client struct {
	typ			int
	mu 			sync.Mutex
	broker 		*Broker
	conn 		net.Conn
	info 		info
	route 		route
	status 		int
	ctx 		context.Context
	cancelFunc 	context.CancelFunc
	session 	*sessions.Session
	subMap 		map[string]*subscription
	subMapMu 	sync.RWMutex
	topicsMgr	*topics.Manager
	subs 		[]interface{}
	qoss 		[]buye
	rmsgs		[]*packets.PublishPacket
	routeSubMap map[string]uint64
	routeSubMapMu	sync.Mutex
	awaitingRel   map[uint16]int64
	maxAwaitingRel int
	inflight 		map[uint16]*inflightElem
	inflightMu 		sync.RWMutex
	mqueue 			*queue.Queue
	retryTimer 		*time.Timer
	retryTimerLock  sync.Mutex

}

type InflightStatus uint8

const (
	Publish InflightStatus = 0
	Pubrel InflightStatus = 1
)

type inflightElem struct {
	status InflightStatus
	packet  *packets.PublishPacket
	timestamp int64
}

type subscription struct {
	client		*client
	topic 		string
	qos 		byte
	share 		bool
	groupName 	string
}

type info struct {
	clientID 	string
	username 	string
	password 	[]byte
	keepalive uint16
	willMsg 	*packets.PublishPacket
	localIP 	string
	remoteIP 	string

}

type route struct {
	remoteID string
	remoteUrl  string
}

var (
	DisconnectedPacket = packets.NewControlPacket(packets.Disconnect).(*packets.DisconnectPacket)
	r					= rand.New(rand.NewSource(time.Now().UnixNano()))
)

func (c *client) init() {
	c.status = Connected
	c.info.localIP,_,_ = net.SplitHostPort(c.conn.LocalAddr().String())
}

