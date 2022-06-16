package main

import (
	"fmt"
	"github.com/qypt15/fmq/dev"
	"log"
	"net"
)


type client struct {
	status   int
	conn 		net.Conn
	localIP 	string
}

func (c *client) tt() {
	c.status = 11
	c.localIP,_,_ = net.SplitHostPort(c.conn.LocalAddr().String())
}


func main() {

	config, err := dev.RetTest("my is test")
	if err != nil {
		log.Fatal("configure broker config error: ", err)
	}
	fmt.Println(config)
	tt := net.Conn.LocalAddr
	fmt.Println(tt)


}