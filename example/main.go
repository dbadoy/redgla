package main

import (
	"fmt"
	"time"

	"github.com/dbadoy/redgla"
)

func main() {
	cfg := redgla.DefaultConfig()
	cfg.Threshold = 5
	cfg.Endpoints = append(cfg.Endpoints, "https://rpc.ankr.com/eth", "https://eth.llamarpc.com", "https://api.securerpc.com/v1")

	redgla, err := redgla.New(redgla.DefaultHeartbeatFn, cfg)
	if err != nil {
		panic(err)
	}

	redgla.Run()

	time.Sleep(3 * time.Second)

	// // Benchmark
	// result, err := redgla.Benchmark(1000, 3)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(result)

	// This request is divided into 5 block request and sent to 2 nodes.
	n1 := time.Now()
	res, err := redgla.BlockByRangeWithBatch(111, 120)
	if err != nil {
		panic(err)
	}

	// 2023-02-24, Seoul: 1.928021s, 10
	fmt.Println(time.Since(n1), len(res))

	n2 := time.Now()
	res, err = redgla.BlockByRange(171, 180)
	if err != nil {
		panic(err)
	}

	// 2023-02-24, Seoul: 3.542532459s, 10
	fmt.Println(time.Since(n2), len(res))
}
