// Copyright (c) 2023, redgla authors <dbadoy4874@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package redgla

import (
	"testing"
)

func TestMakeBatchIndex(t *testing.T) {
	tests := []struct {
		requests int
		clients  int
	}{
		{3221, 5},
		{100, 5},
		{29311, 13},
		{7, 13},
		{100, 30},
	}

	for _, test := range tests {
		indices := makeBatchIndex(test.requests, test.clients)

		totalReq := 0
		for _, index := range indices {
			totalReq += index[1] - index[0]
			if index[1]-index[0] == 0 {
				t.Fatalf("foo")
			}
		}

		if totalReq != test.requests {
			t.Fatalf("invalid makeBatchIndex, want: %d got: %d", test.requests, totalReq)
		}
	}
}

func TestMakeBatchRange(t *testing.T) {
	tests := []struct {
		start   uint64
		end     uint64
		clients int
	}{
		{100, 200, 5},
		{111, 1281, 10},
		{100, 200, 30},
	}

	for _, test := range tests {
		indices := makeBatchRange(test.start, test.end, test.clients)

		totalReq := uint64(0)
		for _, index := range indices {
			totalReq += index[1] - index[0]
			if index[1]-index[0] == 0 {
				t.Fatalf("foo")
			}
		}

		if totalReq != test.end-test.start {
			t.Fatalf("invalid makeBatchRange, want: %d got: %d", test.end-test.start, totalReq)
		}
	}
}
