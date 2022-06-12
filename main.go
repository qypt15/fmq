package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"github.com/qypt15/fmq/broker"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config, err := broker.ConfigureConfig(os.Args[:1])
	if err != nil {
		log.Fatal("configure broker config error: ", err)
	}
	fmt.Println(config)
}