package main

import (
	"fmt"
	"time"

	"github.com/dbadoy/redgla"
)

func main() {
	cfg := redgla.DefaultConfig()
	cfg.Threshold = 5
	cfg.Endpoints = append(cfg.Endpoints, "https://rpc.ankr.com/eth", "https://rpc.flashbots.net")

	redgla, err := redgla.New(redgla.DefaultHeartbeatFn, cfg)
	if err != nil {
		panic(err)
	}

	redgla.Run()

	time.Sleep(3 * time.Second)

	result, err := redgla.Benchmark(1000, 3)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)

	// This request is divided into 5 block request and sent to 2 nodes.
	res, err := redgla.BlockByRangeWithBatch(30, 40)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
