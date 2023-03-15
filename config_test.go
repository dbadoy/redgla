// Copyright (c) 2023, redgla authors <dbadoy4874@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package redgla

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	dcfg := DefaultConfig()

	if dcfg.Threshold != defaultThreshold {
		t.Fatalf("want: %v got: %v", defaultThreshold, dcfg.Threshold)
	}

	if dcfg.HeartbeatInterval != defaultHeartbeatInterval {
		t.Fatalf("want: %v got: %v", defaultHeartbeatInterval, dcfg.HeartbeatInterval)
	}

	if dcfg.HeartbeatTimeout != defaultHeartbeatTimeout {
		t.Fatalf("want: %v got: %v", defaultHeartbeatTimeout, dcfg.HeartbeatTimeout)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		cfg *Config
		err error
	}{
		{
			&Config{
				Endpoints:         []string{"http://127.0.0.1:3821"},
				Threshold:         1,
				RequestTimeout:    defaultRequestTimeout,
				HeartbeatInterval: time.Second,
				HeartbeatTimeout:  time.Second,
			},
			nil,
		},
		{
			&Config{
				Endpoints:         []string{"http://127.0.0.1:3821"},
				Threshold:         100,
				RequestTimeout:    defaultRequestTimeout,
				HeartbeatInterval: 0,
				HeartbeatTimeout:  time.Second,
			},
			errInvalidInterval,
		},
		{
			&Config{
				Endpoints:         []string{"http://127.0.0.1:3821"},
				Threshold:         100,
				RequestTimeout:    defaultRequestTimeout,
				HeartbeatInterval: time.Second,
				HeartbeatTimeout:  0,
			},
			errInvalidTimeout,
		},
		{
			&Config{
				Endpoints:         nil,
				Threshold:         100,
				RequestTimeout:    defaultRequestTimeout,
				HeartbeatInterval: time.Second,
				HeartbeatTimeout:  time.Second,
			},
			errInvalidEndpoint,
		},
		{
			&Config{
				Endpoints:         []string{"http://127.0.0.1:3821", "dbadoy"},
				Threshold:         100,
				RequestTimeout:    defaultRequestTimeout,
				HeartbeatInterval: time.Second,
				HeartbeatTimeout:  time.Second,
			},
			errInvalidEndpoint,
		},
		{
			&Config{
				Endpoints:         []string{"http://127.0.0.1:3821"},
				Threshold:         100,
				RequestTimeout:    0,
				HeartbeatInterval: time.Second,
				HeartbeatTimeout:  time.Second,
			},
			errInvalidTimeout,
		},
		{
			&Config{
				Endpoints:         []string{"ws://127.0.0.1:3821"},
				Threshold:         100,
				RequestTimeout:    defaultRequestTimeout,
				HeartbeatInterval: time.Second,
				HeartbeatTimeout:  time.Second,
			},
			errWebsocketNotSupported,
		},
	}

	for _, test := range tests {
		err := test.cfg.validate()

		if !errors.Is(err, test.err) {
			t.Fatalf("TestConfigValidation, want %v got %v", test.err, err)
		}
	}
}
