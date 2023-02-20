package redgla

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ErrNoAliveNode     = errors.New("there is no alive node")
	ErrTooManyRequests = errors.New("too many requests")
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

	return blockByNumber(clients[0], start, end)
}

func (r *Redgla) BlockByRangeWithBatch(start uint64, end uint64) (map[uint64]*types.Block, error) {
	panic("")
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

	return receiptByTxs(clients[0], txs)
}

func (r *Redgla) ReceiptByTxsWithBatch(txs *types.Transaction) (map[common.Hash]*types.Receipt, error) {
	panic("")
}

func (r *Redgla) dial(endpoints []string) ([]*ethclient.Client, error) {
	res := make([]*ethclient.Client, 0, len(endpoints))

	// There is no acutal TCP connections
	for _, endpoint := range endpoints {
		client, err := ethclient.Dial(endpoint)
		if err != nil {
			return nil, err
		}
		res = append(res, client)
	}

	return res, nil
}

func blockByNumber(node *ethclient.Client, start uint64, end uint64) (map[uint64]*types.Block, error) {
	panic("")
}

func receiptByTxs(node *ethclient.Client, txs []*types.Transaction) (map[common.Hash]*types.Receipt, error) {
	panic("")
}