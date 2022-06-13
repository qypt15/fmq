package broker

import (
	"context"
	"github.com/eapache/queue"
	"github.com/eclipse/paho.mqtt.golang/packets"
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