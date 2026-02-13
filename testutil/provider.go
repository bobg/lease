package testutil

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/bobg/lease"
)

// Provider tests the basic behavior of a [lease.Provider] implementation.
// The provider parameter should be a fresh provider instance for the test.
func Provider(ctx context.Context, t *testing.T, provider lease.Provider) {
	synctest.Run(func() {
		t0 := time.Now()

		secret, err := provider.Acquire(ctx, "test", t0.Add(10*time.Second))
		if err != nil {
			t.Fatalf("Error acquiring lease: %s", err)
		}
		defer provider.Release(ctx, "test", secret)

		_, err = provider.Acquire(ctx, "test", t0.Add(10*time.Second))
		if !errors.Is(err, lease.ErrHeld) {
			t.Errorf("got error %v, want ErrHeld", err)
		}

		secret2, err := provider.Acquire(ctx, "test2", t0.Add(20*time.Second))
		if err != nil {
			t.Fatal(err)
		}
		defer provider.Release(ctx, "test2", secret2)

		time.Sleep(5 * time.Second) // i.e. t0+5s

		_, err = provider.Acquire(ctx, "test", t0.Add(10*time.Second))
		if !errors.Is(err, lease.ErrHeld) {
			t.Errorf("got error %v, want ErrHeld", err)
		}

		time.Sleep(10 * time.Second) // i.e. t0+15s

		secret3, err := provider.Acquire(ctx, "test", t0.Add(40*time.Second))
		if err != nil {
			t.Fatalf("Error acquiring expired lease: %s", err)
		}
		defer provider.Release(ctx, "test", secret3)

		// Can no longer renew the lease with the old secret.
		err = provider.Renew(ctx, "test", secret, t0.Add(20*time.Second))
		if !errors.Is(err, lease.ErrNotHeld) {
			t.Errorf("got error %v, want ErrNotHeld", err)
		}

		_, err = provider.Acquire(ctx, "test2", t0.Add(20*time.Second))
		if !errors.Is(err, lease.ErrHeld) {
			t.Errorf("got error %v, want ErrHeld", err)
		}

		err = provider.Release(ctx, "test2", secret2)
		if err != nil {
			t.Fatalf("Error releasing lease: %s", err)
		}

		time.Sleep(20 * time.Second) // i.e. t0+35s

		err = provider.Renew(ctx, "test", secret3, t0.Add(50*time.Second))
		if err != nil {
			t.Fatalf("Error renewing lease: %s", err)
		}

		time.Sleep(10 * time.Second) // i.e. t0+45s

		_, err = provider.Acquire(ctx, "test", t0.Add(60*time.Second))
		if !errors.Is(err, lease.ErrHeld) {
			t.Errorf("got error %v, want ErrHeld", err)
		}

		time.Sleep(10 * time.Second) // i.e. t0+55s

		secret4, err := provider.Acquire(ctx, "test", t0.Add(80*time.Second))
		if err != nil {
			t.Fatalf("Error acquiring expired lease: %s", err)
		}
		defer provider.Release(ctx, "test", secret4)
	})
}
