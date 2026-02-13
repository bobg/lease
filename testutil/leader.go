package testutil

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/bobg/lease"
)

// Leader tests the ability of a [lease.Provider] implementation to support the [lease.Leader] behavior.
// The provider parameter should be a fresh provider instance for the test.
func Leader(ctx context.Context, t *testing.T, provider lease.Provider) {
	synctest.Test(t, func(t *testing.T) {
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

		synctest.Wait() // Wait for player1 to start and acquire the lease
		<-player1Running

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

		time.Sleep(6 * time.Second) // t0+6s
		synctest.Wait()

		select {
		case <-ctx.Done():
			t.Fatal("Context canceled before seeing whether player 2 must still wait")

		case <-player2Running:
			t.Fatal("Player 2 running too soon")

		case <-time.After(time.Second):
			// ok
		}

		time.Sleep(6 * time.Second) // t0+12s
		synctest.Wait()

		select {
		case <-ctx.Done():
			t.Fatal("Context canceled before seeing whether player 2 must still wait")

		case <-player2Running:
			t.Fatal("Player 2 running too soon")

		case <-time.After(time.Second):
			// ok
		}

		close(player1Exit)
		synctest.Wait()
		<-player1Done

		time.Sleep(12 * time.Second) // t0+24s
		synctest.Wait()

		<-player2Running

		if !player1Ran {
			t.Fatalf("player 1 ran = %v, want true", player1Ran)
		}

		var cberr lease.CallbackError
		if !errors.As(player1Err, &cberr) {
			t.Fatalf("player 1 error = %v, want lease.CallbackError", player1Err)
		}

		if !errors.Is(cberr, player1WantErr) {
			t.Fatalf("player 1 error = %v, want %v", player1Err, player1WantErr)
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

		time.Sleep(6 * time.Second) // t0+30s
		synctest.Wait()

		cancel()
		synctest.Wait()

		<-player2Done
		<-player3Done

		if !player2Ran {
			t.Fatal("player 2 ran = false, want true")
		}
		if !errors.Is(player2Err, context.Canceled) {
			t.Fatalf("player 2 error = %v, want context.Canceled", player2Err)
		}

		if player3Ran {
			t.Fatal("player 3 ran = true, want false")
		}
		if errors.As(player3Err, &cberr) {
			t.Fatal("player 3 error is a lease.CallbackError but should not be")
		}
		if !errors.Is(player3Err, context.Canceled) {
			t.Fatalf("player 3 error = %v, want context.Canceled", player3Err)
		}
	})
}
