package broker

import (
	"context"
	"net"
	"regexp"
	"sync"
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


}