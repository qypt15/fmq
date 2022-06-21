package broker

import (
	"crypto/tls"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/qypt15/fmq/broker/lib/sessions"
	"github.com/qypt15/fmq/broker/lib/topics"
	"github.com/qypt15/fmq/plugins/auth"
	"github.com/qypt15/fmq/plugins/bridge"
	"github.com/qypt15/fmq/pool"
	"sync"
)

const (
	MessagePoolNum        = 1024
	MessagePoolMessageNum = 1024
)

type Message struct {
	client *client
	packet packets.ControlPacket
}

type Broker struct {
	id          string
	mu          sync.Mutex
	config      *Config
	tlsConfig   *tls.Config
	wpool       *pool.WorkerPool
	clients     sync.Map
	routes      sync.Map
	remotes     sync.Map
	nodes       map[string]interface{}
	clusterPool chan *Message
	topicsMgr   *topics.Manager
	sessionMgr  *sessions.Manager
	auth        auth.Auth
	bridgeMQ    bridge.BridgeMQ
}

func (b *Broker) SubmitWork(clientId string, msg *Message) {
	if b.wpool == nil {
		b.wpool = pool.New(b.config.Worker)
	}
	if msg.client.typ == CLUSTER {
		b.clusterPool <- msg
	} else {
		b.wpool.Submit(clientId, func() {
			ProcessMessage(msg)
		})
	}
}
