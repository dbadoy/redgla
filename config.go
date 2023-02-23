package redgla

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

const (
	defaultThreshold         = 100
	defaultRequestTimeout    = 5 * time.Minute
	defaultHeartbeatInterval = 3 * time.Second
	defaultHeartbeatTimeout  = time.Second
)

var (
	errInvalidInterval = errors.New("invalid heartbeat interval")
	errInvalidTimeout  = errors.New("invalid heartbeat timeout")
	errInvalidEndpoint = errors.New("invalid endpoint")
)

type Config struct {
	// A list of endpoints to send batch requests to.
	Endpoints []string

	// Threshold to send a batch request. If the number of requests is
	// greater than the value, they are converted to batch requests.
	Threshold int

	RequestTimeout time.Duration

	// Ping interval for checks alive endpoints.
	HeartbeatInterval time.Duration

	// Timeout for requests to determine 'alive'.
	HeartbeatTimeout time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		Threshold:         defaultThreshold,
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
