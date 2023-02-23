package redgla

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Internal messages.
// temp

type msg struct {
	endpoint string
	err      error
	v        interface{}
}

func (m *msg) benchmarkResponse() time.Duration {
	return m.v.(time.Duration)
}

func (m *msg) blockResponse() map[uint64]*types.Block {
	return m.v.(map[uint64]*types.Block)
}

func (m *msg) receiptResponse() map[common.Hash]*types.Receipt {
	return m.v.(map[common.Hash]*types.Receipt)
}
