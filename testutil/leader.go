package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/bobg/lease"
)

// Leader tests the ability of a [lease.Provider] implementation to support the [lease.Leader] behavior.
func Leader(ctx context.Context, tb testing.TB, factory Factory) {
	var (
		mockClock = clock.NewMock()
		t0        = time.Date(1977, 8, 5, 0, 0, 0, 0, time.UTC)
	)

	mockClock.Set(t0)

	provider, err := factory(mockClock)
	if err != nil {
		tb.Fatal(err)
	}

	leader := lease.Leader{
		Name:  "leader-test",
		Dur:   10 * time.Second,
		Retry: 10 * time.Second,
		Renew: 5 * time.Second,
	}

	var (
		player1Running = make(chan struct{})
		player1Exit    = make(chan struct{})
		player1Err     error
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		_, player1Err = leader.Run(ctx, provider, func(context.Context) error {
			close(player1Running)
			<-player1Exit
			return nil
		})
	}()

	select {
	case <-ctx.Done():
		tb.Fatal("Context canceled before player 1 could run")

	case <-player1Running:
	}

	// Player 1 is running.

	player2Running := make(chan struct{})

	go func() {
		leader.Run(ctx, provider, func(context.Context) error {
			close(player2Running)
			return nil
		})
	}()

	mockClock.Add(6 * time.Second) // t0+6s

	select {
	case <-ctx.Done():
		tb.Fatal("Context canceled before seeing whether player 2 must still wait")

	case <-player2Running:
		tb.Fatal("Player 2 running too soon")

	case <-time.After(time.Second):
		// ok
	}

	mockClock.Add(6 * time.Second) // t0+12s

	select {
	case <-ctx.Done():
		tb.Fatal("Context canceled before seeing whether player 2 must still wait")

	case <-player2Running:
		tb.Fatal("Player 2 running too soon")

	case <-time.After(time.Second):
		// ok
	}

	close(player1Exit)

	mockClock.Add(12 * time.Second) // t0+24s

	select {
	case <-ctx.Done():
		tb.Fatal("Context canceled before player 2 could run")

	case <-player2Running:
		// ok
	}

	if player1Err != nil {
		tb.Fatal(player1Err)
	}
}
