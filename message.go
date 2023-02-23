package redgla

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Internal messages.

type benchmarkMsg struct {
	endpoint string
	err      error
	spent    time.Duration
}

type blockMsg struct {
	endpoint string
	err      error
	res      map[uint64]*types.Block
}

type receiptMsg struct {
	endpoint string
	err      error
	res      map[common.Hash]*types.Receipt
}
