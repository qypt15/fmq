package main

import (
	"fmt"
	"log"
	"github.com/qypt15/fmq/dev"
)

func main() {

	config, err := dev.RetTest("my is test")
	if err != nil {
		log.Fatal("configure broker config error: ", err)
	}
	fmt.Println(config)



}