package redgla

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ErrNoAliveNode     = errors.New("there is no alive node")
	ErrTooManyRequests = errors.New("too many requests. use *Batch methods")
)

type Redgla struct {
	list *beater
	cfg  *Config
}

func (r *Redgla) BlockByRange(start uint64, end uint64) (map[uint64]*types.Block, error) {
	nodes := r.list.liveNodes()
	if len(nodes) == 0 {
		return nil, ErrNoAliveNode
	}

	clients, err := r.dial([]string{nodes[0]})
	if err != nil {
		return nil, err
	}

	if r.cfg.Threshold < int(end-start) {
		return nil, ErrTooManyRequests
	}

	return blockByNumber(clients[0], start, end, r.cfg.RequestTimeout)
}

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
		resc   = make(chan map[uint64]*types.Block, len(nodes))
		result = make(map[uint64]*types.Block, end-start)
	)

	ranges := makeBatchRange(start, end, len(nodes))
	for i, rg := range ranges {
		go func(client *ethclient.Client, start uint64, end uint64) {
			r, err := blockByNumber(client, start, end, r.cfg.RequestTimeout)
			if err != nil {
				resc <- nil
				return
			}
			resc <- r
		}(clients[i], rg[0], rg[1])
	}

	for i := 0; i < cap(resc); i++ {
		res := <-resc
		if res == nil {
			return nil, errors.New("request failed during batch operation")
		}

		for k, v := range res {
			result[k] = v
		}
	}

	return result, nil
}

func (r *Redgla) ReceiptByTxs(txs []*types.Transaction) (map[common.Hash]*types.Receipt, error) {
	nodes := r.list.liveNodes()
	if len(nodes) == 0 {
		return nil, ErrNoAliveNode
	}

	clients, err := r.dial([]string{nodes[0]})
	if err != nil {
		return nil, err
	}

	if r.cfg.Threshold < len(txs) {
		return nil, ErrTooManyRequests
	}

	return receiptByTxs(clients[0], txs, r.cfg.RequestTimeout)
}

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
		resc   = make(chan map[common.Hash]*types.Receipt, len(nodes))
		result = make(map[common.Hash]*types.Receipt, len(txs))
	)

	indices := makeBatchIndex(len(txs), len(nodes))
	for i, index := range indices {
		go func(client *ethclient.Client, txs []*types.Transaction) {
			r, err := receiptByTxs(client, txs, r.cfg.RequestTimeout)
			if err != nil {
				resc <- nil
				return
			}
			resc <- r
		}(clients[i], txs[index[0]:index[1]])
	}

	for i := 0; i < cap(resc); i++ {
		res := <-resc
		if res == nil {
			return nil, errors.New("request failed during batch operation")
		}

		for k, v := range res {
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

func blockByNumber(client *ethclient.Client, start uint64, end uint64, timeout time.Duration) (res map[uint64]*types.Block, err error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	res = make(map[uint64]*types.Block, end-start)

	for ; start <= end; start++ {
		select {
		case <-timer.C:
			return nil, errors.New("timeout")
		default:
		}

		res[start], err = client.BlockByNumber(context.Background(), big.NewInt(int64(start)))
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func receiptByTxs(client *ethclient.Client, txs []*types.Transaction, timeout time.Duration) (res map[common.Hash]*types.Receipt, err error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	res = make(map[common.Hash]*types.Receipt, len(txs))

	for _, tx := range txs {
		select {
		case <-timer.C:
			return nil, errors.New("timeout")
		default:
		}

		res[tx.Hash()], err = client.TransactionReceipt(context.Background(), tx.Hash())
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
