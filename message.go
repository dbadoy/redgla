// Copyright (c) 2023, redgla authors <dbadoy4874@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package redgla

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Internal messages.

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

func (m *msg) transactionResponse() map[common.Hash]*types.Transaction {
	return m.v.(map[common.Hash]*types.Transaction)
}

func (m *msg) receiptResponse() map[common.Hash]*types.Receipt {
	return m.v.(map[common.Hash]*types.Receipt)
}
