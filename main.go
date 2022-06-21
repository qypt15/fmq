package main

import (
	"fmt"
	"github.com/qypt15/fmq/broker"
	"log"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// config, err := broker.ConfigureConfig(os.Args[:1])
	config, err := broker.Test(os.Args[:1])

	if err != nil {
		log.Fatal("configure broker config error: ", err)
	}

	fmt.Println(config)
}
