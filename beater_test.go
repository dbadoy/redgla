package redgla

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultHeartbeatFn(t *testing.T) {
	tests := []struct {
		endpoint string
		succeed  bool
	}{
		{
			"dbadoy",
			false,
		},
		{
			"https://rpc.ankr.com/eth",
			true,
		},
		{
			"https://rpc.flashbots.net",
			true,
		},
	}

	// Is context timeout works?
	for _, test := range tests {
		// Always return timeout error.
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		if err := DefaultHeartbeatFn(ctx, test.endpoint); err == nil {
			t.Fatalf("want: %v got: %v", test.succeed, false)
		}
	}

	for _, test := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := DefaultHeartbeatFn(ctx, test.endpoint)
		if err == nil && !test.succeed {
			t.Fatalf("want: %v got: %v", test.succeed, err == nil)
		}
		if err != nil && test.succeed {
			t.Fatalf("want: %v got: %v", test.succeed, err == nil)
		}
	}
}

func TestBeat(t *testing.T) {
	// Always success.
	fn := func(context.Context, string) error {
		return nil
	}

	beater, err := newBeater("test", []string{"http://127.0.0.1:1823", "http://127.0.0.1:1824"}, fn, time.Second, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	liveNodes := beater.beat(beater.endpoints)
	if len(liveNodes) != 2 {
		t.Fatalf("beater.beat failure, want: %v got: %v", 2, len(beater.liveNodes()))
	}
}

func TestBeaterWithAlwaysSuccess(t *testing.T) {
	// Always success.
	fn := func(context.Context, string) error {
		return nil
	}

	beater, err := newBeater("test", []string{"http://127.0.0.1:1823", "http://127.0.0.1:1824"}, fn, time.Second, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	beater.run()

	time.Sleep(10 * time.Millisecond)

	if len(beater.liveNodes()) != 2 {
		t.Fatalf("beater.beat failure, want: %v got: %v", 2, len(beater.liveNodes()))
	}

	beater.stop()

	if len(beater.liveNodes()) != 0 {
		t.Fatalf("beater.beat failure, want: %v got: %v", 0, len(beater.liveNodes()))
	}
}

func TestBeaterWithAlwaysFail(t *testing.T) {
	// Always success.
	fn := func(context.Context, string) error {
		return errors.New("no")
	}

	beater, err := newBeater("test", []string{"http://127.0.0.1:1823", "http://127.0.0.1:1824"}, fn, time.Second, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	beater.run()

	time.Sleep(10 * time.Millisecond)

	if len(beater.liveNodes()) != 0 {
		t.Fatalf("beater.beat failure, want: %v got: %v", 2, len(beater.liveNodes()))
	}

	beater.stop()

	if len(beater.liveNodes()) != 0 {
		t.Fatalf("beater.beat failure, want: %v got: %v", 0, len(beater.liveNodes()))
	}
}
