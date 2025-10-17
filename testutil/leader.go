package testutil

import (
	"context"
	"errors"
	"fmt"
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
		player1Running = make(chan struct{}) // closes when player1 starts
		player1Exit    = make(chan struct{}) // close this to make player1 exit
		player1Done    = make(chan struct{}) // closes after player1 exits
		player1Ran     bool
		player1Err     error
		player1WantErr = fmt.Errorf("player 1 error")
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		player1Ran, player1Err = leader.Run(ctx, provider, func(context.Context) error {
			close(player1Running)
			<-player1Exit
			return player1WantErr
		})
		close(player1Done)
	}()

	select {
	case <-ctx.Done():
		tb.Fatal("Context canceled before player 1 could run")

	case <-player1Running:
	}

	// Player 1 is running.

	var (
		player2Running = make(chan struct{})
		player2Done    = make(chan struct{})
		player2Ran     bool
		player2Err     error
	)

	go func() {
		player2Ran, player2Err = leader.Run(ctx, provider, func(ctx context.Context) error {
			close(player2Running)
			<-ctx.Done()
			return ctx.Err()
		})
		close(player2Done)
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
	<-player1Done

	mockClock.Add(12 * time.Second) // t0+24s

	select {
	case <-ctx.Done():
		tb.Fatal("Context canceled before player 2 could run")

	case <-player2Running:
		// ok
	}

	if !player1Ran {
		tb.Fatalf("player 1 ran = %v, want true", player1Ran)
	}

	var cberr lease.CallbackError
	if !errors.As(player1Err, &cberr) {
		tb.Fatalf("player 1 error = %v, want lease.CallbackError", player1Err)
	}

	if !errors.Is(cberr, player1WantErr) {
		tb.Fatalf("player 1 error = %v, want %v", player1Err, player1WantErr)
	}

	var (
		player3Running = make(chan struct{})
		player3Done    = make(chan struct{})
		player3Ran     bool
		player3Err     error
	)

	go func() {
		player3Ran, player3Err = leader.Run(ctx, provider, func(context.Context) error {
			close(player3Running)
			return nil
		})
		close(player3Done)
	}()

	mockClock.Add(6 * time.Second) // t0+30s

	cancel()

	<-player2Done
	<-player3Done

	if !player2Ran {
		tb.Fatal("player 2 ran = false, want true")
	}
	if !errors.Is(player2Err, context.Canceled) {
		tb.Fatalf("player 2 error = %v, want context.Canceled", player2Err)
	}

	if player3Ran {
		tb.Fatal("player 3 ran = true, want false")
	}
	if errors.As(player3Err, &cberr) {
		tb.Fatal("player 3 error is a lease.CallbackError but should not be")
	}
	if !errors.Is(player3Err, context.Canceled) {
		tb.Fatalf("player 3 error = %v, want context.Canceled", player3Err)
	}
}
