// Copyright (c) 2023, redgla authors <dbadoy4874@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package redgla

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

const (
	defaultThreshold         = 100
	defaultRequestTimeout    = 30 * time.Minute // See Config.RequestTimeout comment.
	defaultHeartbeatInterval = 3 * time.Second
	defaultHeartbeatTimeout  = time.Second
)

var (
	errInvalidEndpoint = errors.New("invalid endpoint")
	errInvalidInterval = errors.New("invalid heartbeat interval")
	errInvalidTimeout  = errors.New("invalid timeout")
)

type Config struct {
	// A list of endpoints to send batch requests to.
	Endpoints []string

	// Threshold to send a batch request. If the number of requests is
	// greater than the value, they are converted to batch requests.
	Threshold int

	// This is the timeout of the request to the Ethereum node. It also
	// seems okay to give a very large value and rely on the Ethereum
	// node's request timeout.
	RequestTimeout time.Duration

	// Ping interval for checks alive endpoints.
	HeartbeatInterval time.Duration

	// Timeout for requests to determine 'alive'.
	HeartbeatTimeout time.Duration
}

// There is no default value for Endpoints. Set the Endpoints
// on the created default Config.
func DefaultConfig() *Config {
	return &Config{
		Threshold:         defaultThreshold,
		RequestTimeout:    defaultRequestTimeout,
		HeartbeatInterval: defaultHeartbeatInterval,
		HeartbeatTimeout:  defaultHeartbeatTimeout,
	}
}

func (c *Config) validate() error {
	if len(c.Endpoints) == 0 {
		return errInvalidEndpoint
	}

	for _, endpoint := range c.Endpoints {
		if err := validateEndpoint(endpoint); err != nil {
			return err
		}
	}

	if c.RequestTimeout == 0 {
		return errInvalidTimeout
	}

	if c.HeartbeatInterval == 0 {
		return errInvalidInterval
	}

	if c.HeartbeatTimeout == 0 {
		return errInvalidTimeout
	}

	return nil
}

func validateEndpoint(endpoint string) error {
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return fmt.Errorf("%s: %w", endpoint, errInvalidEndpoint)
	}

	return nil
}
