package broker

import (
	"crypto/tls"
	"github.com/qypt15/fmq/bridge"
	"github.com/qypt15/fmq/plugins/auth"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"sync"
)

const (
	MessagePoolNum 		= 1024
	MessagePoolMessageNum = 1024
)

type Message struct {
	client *client
	packet packets.ControlPacket
}

type Broker struct {
	id 		string
	mu 		sync.Mutex
	config  *Config
	tlsConfig *tls.Config
	wpool    *pool.WorkerPool
	clients  sync.Map
	routes  sync.Map
	remotes   sync.Map
	nodes 		map[string]interface{}
	clusterPool 	chan  *Message
	topicsMgr 	*topics.Manager
	sessionMgr  *sessions.Manager
	auth 	auth.Auth
	bridgeMQ 	bridge.BridgeMQ
}