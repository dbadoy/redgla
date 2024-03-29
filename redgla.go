// Copyright (c) 2023, redgla authors <dbadoy4874@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package redgla

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ErrNoAliveNode  = errors.New("there is no alive node")
	ErrBatchFailure = errors.New("batch request failure")
)

type Redgla struct {
	isRun uint32

	list *beater
	cfg  *Config
}

func New(fn HeartbeatFn, cfg *Config) (*Redgla, error) {
	if fn == nil {
		fn = DefaultHeartbeatFn
	}

	if cfg == nil {
		return nil, errors.New("Config must not be nil")
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	beater, err := newBeater("beater", cfg.Endpoints, fn, cfg.HeartbeatInterval, cfg.HeartbeatTimeout)
	if err != nil {
		return nil, err
	}

	return &Redgla{0, beater, cfg}, nil
}

func (r *Redgla) Run() {
	if atomic.LoadUint32(&r.isRun) == 0 {
		r.list.run()
		atomic.StoreUint32(&r.isRun, 1)
	}
}

func (r *Redgla) Stop() {
	if atomic.LoadUint32(&r.isRun) == 1 {
		r.list.stop()
		atomic.StoreUint32(&r.isRun, 0)
	}
}

// AddNode adds the target endpoint to the list of batch processing nodes.
// The endpoint entered will take effect starting from the next
// HeartbeatInterval.
func (r *Redgla) AddNode(endpoint string) error {
	return r.list.add(endpoint)
}

// DelNode removes the target endpoint from the list of batch processing
// nodes. Removed endpoints take effect from the next HeartbeatInterval.
func (r *Redgla) DelNode(endpoint string) error {
	return r.list.delete(endpoint)
}

// Benchmark measures and returns the response time of each node.
//
// Batch request performance is matched to the speed of the slowest node.
// Removing nodes that are too slow to respond from the list can help
// improve performance.
// This method performs a benchmark for requests to fetch 'cnt' times a
// random number of block numbers less than 'height'.
func (r *Redgla) Benchmark(height uint64, cnt int) (map[string]time.Duration, error) {
	nodes := r.list.liveNodes()
	if len(nodes) == 0 {
		return nil, ErrNoAliveNode
	}

	clients, err := r.dial(nodes)
	if err != nil {
		return nil, err
	}

	var (
		resc   = make(chan *msg, len(nodes))
		result = make(map[string]time.Duration)
	)

	for i := 0; i < len(clients); i++ {
		go func(client *ethclient.Client, endpoint string) {
			start := time.Now()
			// Requesting for the same block number results in a faster
			// response time (it seems to be cached), so we ask for a
			// random number.
			randBN := rand.Int63n(int64(height-1) + 1)
			for i := 0; i < cnt; i++ {
				_, err := client.BlockByNumber(context.Background(), big.NewInt(randBN))
				if err != nil {
					resc <- &msg{endpoint, err, 0}
					return
				}
			}
			resc <- &msg{endpoint, nil, time.Since(start)}
		}(clients[i], nodes[i])
	}

	for i := 0; i < cap(resc); i++ {
		res := <-resc
		if res.err != nil {
			return nil, fmt.Errorf("%w: %s (%s)", res.err, res.endpoint, "some request failed during benchmark")
		}
		result[res.endpoint] = res.benchmarkResponse()
	}

	return result, nil
}

// BlockByRange requests blocks from a range to a node.
func (r *Redgla) BlockByRange(start uint64, end uint64) (map[uint64]*types.Block, error) {
	nodes := r.list.liveNodes()
	if len(nodes) == 0 {
		return nil, ErrNoAliveNode
	}

	clients, err := r.dial([]string{nodes[0]})
	if err != nil {
		return nil, err
	}

	return blockByRange(clients[0], start, end, r.cfg.RequestTimeout, nil)
}

// BlockByRangeWithBatch transmits and receives batch requests to
// healthy nodes among the list of registered nodes.
func (r *Redgla) BlockByRangeWithBatch(start uint64, end uint64) (map[uint64]*types.Block, error) {
	if r.cfg.Threshold >= int(end-start) {
		return r.BlockByRange(start, end)
	}

	nodes := r.list.liveNodes()
	if len(nodes) == 0 {
		return nil, ErrNoAliveNode
	}

	clients, err := r.dial(nodes)
	if err != nil {
		return nil, err
	}

	var (
		resc   = make(chan *msg, len(nodes))
		result = make(map[uint64]*types.Block, end-start)
	)

	quit := make(chan struct{})
	ranges := makeBatchRange(start, end, len(nodes))
	for i, rg := range ranges {
		go func(client *ethclient.Client, endpoint string, start uint64, end uint64) {
			r, err := blockByRange(client, start, end, r.cfg.RequestTimeout, quit)
			if err != nil {
				resc <- &msg{endpoint, err, nil}
				return
			}
			resc <- &msg{endpoint, nil, r}
		}(clients[i], nodes[i], rg[0], rg[1])
	}

	for i := 0; i < cap(resc); i++ {
		res := <-resc
		if res.err != nil {
			close(quit)
			return nil, fmt.Errorf("%w: %s (%s)", res.err, res.endpoint, "request failed during batch operation")
		}

		for k, v := range res.blockResponse() {
			result[k] = v
		}
	}

	return result, nil
}

// TransactionByHashes requests transactions from given hashes to a node.
func (r *Redgla) TransactionByHashes(hashes []common.Hash) (map[common.Hash]*types.Transaction, error) {
	nodes := r.list.liveNodes()
	if len(nodes) == 0 {
		return nil, ErrNoAliveNode
	}

	clients, err := r.dial([]string{nodes[0]})
	if err != nil {
		return nil, err
	}

	return transactionByHashes(clients[0], hashes, r.cfg.RequestTimeout, nil)
}

// TransactionByHashesWithBatch transmits and receives batch requests to
// healthy nodes among the list of registered nodes.
func (r *Redgla) TransactionByHashesWithBatch(hashes []common.Hash) (map[common.Hash]*types.Transaction, error) {
	if r.cfg.Threshold >= len(hashes) {
		return r.TransactionByHashes(hashes)
	}

	nodes := r.list.liveNodes()
	if len(nodes) == 0 {
		return nil, ErrNoAliveNode
	}

	clients, err := r.dial(nodes)
	if err != nil {
		return nil, err
	}

	var (
		resc   = make(chan *msg, len(nodes))
		result = make(map[common.Hash]*types.Transaction, len(hashes))
	)

	quit := make(chan struct{})
	indices := makeBatchIndex(len(hashes), len(nodes))
	for i, index := range indices {
		go func(client *ethclient.Client, endpoint string, hashes []common.Hash) {
			r, err := transactionByHashes(client, hashes, r.cfg.RequestTimeout, quit)
			if err != nil {
				resc <- &msg{endpoint, err, nil}
				return
			}
			resc <- &msg{endpoint, nil, r}
		}(clients[i], nodes[i], hashes[index[0]:index[1]])
	}

	for i := 0; i < cap(resc); i++ {
		res := <-resc
		if res.err != nil {
			close(quit)
			return nil, fmt.Errorf("%w: %s (%s)", res.err, res.endpoint, "request failed during batch operation")
		}

		for k, v := range res.transactionResponse() {
			result[k] = v
		}
	}

	return result, nil
}

// ReceiptByTxs requests receipts from given transactions to a node.
func (r *Redgla) ReceiptByTxs(txs []*types.Transaction) (map[common.Hash]*types.Receipt, error) {
	nodes := r.list.liveNodes()
	if len(nodes) == 0 {
		return nil, ErrNoAliveNode
	}

	clients, err := r.dial([]string{nodes[0]})
	if err != nil {
		return nil, err
	}

	return receiptByTxs(clients[0], txs, r.cfg.RequestTimeout, nil)
}

// ReceiptByTxsWithBatch transmits and receives batch requests to
// healthy nodes among the list of registered nodes.
func (r *Redgla) ReceiptByTxsWithBatch(txs []*types.Transaction) (map[common.Hash]*types.Receipt, error) {
	if r.cfg.Threshold >= len(txs) {
		return r.ReceiptByTxs(txs)
	}

	nodes := r.list.liveNodes()
	if len(nodes) == 0 {
		return nil, ErrNoAliveNode
	}

	clients, err := r.dial(nodes)
	if err != nil {
		return nil, err
	}

	var (
		resc   = make(chan *msg, len(nodes))
		result = make(map[common.Hash]*types.Receipt, len(txs))
	)

	quit := make(chan struct{})
	indices := makeBatchIndex(len(txs), len(nodes))
	for i, index := range indices {
		go func(client *ethclient.Client, endpoint string, txs []*types.Transaction) {
			r, err := receiptByTxs(client, txs, r.cfg.RequestTimeout, quit)
			if err != nil {
				resc <- &msg{endpoint, err, nil}
				return
			}
			resc <- &msg{endpoint, nil, r}
		}(clients[i], nodes[i], txs[index[0]:index[1]])
	}

	for i := 0; i < cap(resc); i++ {
		res := <-resc
		if res.err != nil {
			close(quit)
			return nil, fmt.Errorf("%w: %s (%s)", res.err, res.endpoint, "request failed during batch operation")
		}

		for k, v := range res.receiptResponse() {
			result[k] = v
		}
	}

	return result, nil
}

func (r *Redgla) dial(endpoints []string) ([]*ethclient.Client, error) {
	res := make([]*ethclient.Client, 0, len(endpoints))

	// All of them are dialed and returned even if they are not used.
	// It's seems OK because no actual communication with the node.
	for _, endpoint := range endpoints {
		client, err := ethclient.Dial(endpoint)
		if err != nil {
			return nil, err
		}
		res = append(res, client)
	}

	return res, nil
}

// quit: A Channel that stops all goroutine execution if any of the
//       batch requests fail. If the stop logic of the goroutine is
//       not required, it is nil (i.e. a single request).

func blockByRange(client *ethclient.Client, start uint64, end uint64, timeout time.Duration, quit chan struct{}) (res map[uint64]*types.Block, err error) {
	res = make(map[uint64]*types.Block, end-start)

	if quit == nil {
		quit = make(chan struct{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for ; start <= end; start++ {
		select {
		case _, ok := <-quit:
			if !ok {
				return nil, ErrBatchFailure
			}
		default:
		}

		res[start], err = client.BlockByNumber(ctx, big.NewInt(int64(start)))
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func transactionByHashes(client *ethclient.Client, hashes []common.Hash, timeout time.Duration, quit chan struct{}) (res map[common.Hash]*types.Transaction, err error) {
	res = make(map[common.Hash]*types.Transaction, len(hashes))

	if quit == nil {
		quit = make(chan struct{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, hash := range hashes {
		select {
		case _, ok := <-quit:
			if !ok {
				return nil, ErrBatchFailure
			}
		default:
		}

		res[hash], _, err = client.TransactionByHash(ctx, hash)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func receiptByTxs(client *ethclient.Client, txs []*types.Transaction, timeout time.Duration, quit chan struct{}) (res map[common.Hash]*types.Receipt, err error) {
	res = make(map[common.Hash]*types.Receipt, len(txs))

	if quit == nil {
		quit = make(chan struct{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, tx := range txs {
		select {
		case _, ok := <-quit:
			if !ok {
				return nil, ErrBatchFailure
			}
		default:
		}

		res[tx.Hash()], err = client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func makeBatchIndex(requests int, clients int) [][2]int {
	r := make([][2]int, 0)

	batch := requests/clients + 1

	accum := 0
	for i := 0; i < clients; i++ {
		next := accum + batch
		if next > requests {
			next = requests
		}
		m := [2]int{accum, next}
		accum = next

		r = append(r, m)
		if accum == requests {
			return r
		}
	}

	return r
}

func makeBatchRange(start uint64, end uint64, clients int) [][2]uint64 {
	r := make([][2]uint64, 0)

	batch := (end-start)/uint64(clients) + 1

	for i := 0; i < clients; i++ {
		next := start + batch
		if next > end {
			next = end
		}
		m := [2]uint64{start, next}
		start = next

		r = append(r, m)
		if start == end {
			return r
		}
	}

	return r
}
