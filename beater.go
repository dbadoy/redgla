// Copyright (c) 2023, redgla authors <dbadoy4874@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package redgla

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// HeartbeatFn is a method that can check whether the endpoint is working
// or not. Since it is not possible to specify which service endpoint it
// is, it is appropriately injected from the outside according to the usage.
type HeartbeatFn func(ctx context.Context, endpoint string) error

func DefaultHeartbeatFn(ctx context.Context, endpoint string) error {
	// context.WithTimeout has no effect on DialContext.
	client, err := ethclient.DialContext(ctx, endpoint)
	if err != nil {
		return err
	}

	// It is recommended to make at least one rpc call.
	_, err = client.ChainID(ctx)
	return err
}

// Beater manages the status list by examining whether the endpoints
// registered in the list are operating normally. The endpoint should be
// URL format.
type beater struct {
	name      string
	endpoints []string

	mu sync.RWMutex

	// Sort the members in order of fastest response time.
	//
	// https://github.com/dbadoy/redgla/pull/3
	members priorityQueue

	quit chan struct{}

	fn HeartbeatFn

	interval time.Duration
	timeout  time.Duration
}

type message struct {
	endpoint string
	spent    time.Duration
}

func newBeater(name string, endpoints []string, fn HeartbeatFn, interval, timeout time.Duration) (*beater, error) {
	for _, endpoint := range endpoints {
		if err := isValidEndpoint(endpoint); err != nil {
			return nil, err
		}
	}

	return &beater{
		name:      name,
		endpoints: endpoints,
		quit:      make(chan struct{}),
		fn:        fn,
		interval:  interval,
		timeout:   timeout,
	}, nil
}

func (b *beater) run() {
	go b.loop()
}

func (b *beater) stop() {
	b.quit <- struct{}{}
	b.members = make(priorityQueue, 0)
}

func (b *beater) loop() {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			var (
				result = b.beat(b.endpoints)
				heap   = make(priorityQueue, 0)
			)

			for member, spent := range result {
				heap.add(member, spent)
			}

			b.mu.Lock()
			// TODO(dbadoy): We need to find a better way than reallocating every time.
			b.members = heap
			b.mu.Unlock()

			timer.Reset(b.interval)

		case <-b.quit:
			return
		}
	}
}

func (b *beater) beat(endpoints []string) map[string]time.Duration {
	resc := make(chan *message, len(endpoints))

	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	start := time.Now()
	for _, endpoint := range endpoints {
		go func(t string) {
			if err := b.fn(ctx, t); err != nil {
				resc <- nil
				return
			}
			resc <- &message{t, time.Since(start)}
		}(endpoint)
	}

	m := make(map[string]time.Duration)

	for i := 0; i < cap(resc); i++ {
		msg := <-resc
		if msg != nil {
			m[msg.endpoint] = msg.spent
		}
	}

	return m
}

func (b *beater) add(endpoint string) error {
	if err := isValidEndpoint(endpoint); err != nil {
		return err
	}

	for _, node := range b.nodes() {
		if node == endpoint {
			return errors.New("already exist")
		}
	}

	b.mu.Lock()
	b.endpoints = append(b.endpoints, endpoint)
	b.mu.Unlock()

	return nil
}

func (b *beater) delete(endpoint string) error {
	if err := isValidEndpoint(endpoint); err != nil {
		return err
	}

	for i, node := range b.nodes() {
		// We don't need to delete a node and remove it
		// from 'b.members'; add/delete nodes means
		// applying them in the next 'p.beat'.
		if node == endpoint {
			b.mu.Lock()
			b.endpoints[i] = b.endpoints[len(b.endpoints)-1]
			b.endpoints = b.endpoints[:len(b.endpoints)-1]
			b.mu.Unlock()

			return nil
		}
	}

	return errors.New("not exist endpoint")
}

func (b *beater) nodes() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.endpoints
}

// The result isn't fully sorted, but it's clear that
// the first value is the highest priority. What we
// want is the fastest first item, so we just use it.
func (b *beater) liveNodes() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.members.keys()
}
