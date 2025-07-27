package testutil

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/bobg/lease"
)

// Factory creates a [lease.Provider] using the given [lease.Clock].
type Factory func(lease.Clock) (lease.Provider, error)

// Provider tests the basic behavior of a [lease.Provider] implementation.
func Provider(ctx context.Context, tb testing.TB, factory Factory) {
	tb.Helper()

	var (
		mockClock = clock.NewMock()
		t0        = time.Date(1977, 8, 5, 0, 0, 0, 0, time.UTC)
	)
	mockClock.Set(t0)

	provider, err := factory(mockClock)
	if err != nil {
		tb.Fatal(err)
	}

	secret, err := provider.Acquire(ctx, "test", t0.Add(10*time.Second))
	if err != nil {
		tb.Fatalf("Error acquiring lease: %s", err)
	}
	defer provider.Release(ctx, "test", secret)

	_, err = provider.Acquire(ctx, "test", t0.Add(10*time.Second))
	if !errors.Is(err, lease.ErrHeld) {
		tb.Errorf("got error %v, want ErrHeld", err)
	}

	secret2, err := provider.Acquire(ctx, "test2", t0.Add(20*time.Second))
	if err != nil {
		tb.Fatal(err)
	}
	defer provider.Release(ctx, "test2", secret2)

	mockClock.Add(5 * time.Second) // i.e. t0+5s

	_, err = provider.Acquire(ctx, "test", t0.Add(10*time.Second))
	if !errors.Is(err, lease.ErrHeld) {
		tb.Errorf("got error %v, want ErrHeld", err)
	}

	mockClock.Add(10 * time.Second) // i.e. t0+15s

	secret3, err := provider.Acquire(ctx, "test", t0.Add(40*time.Second))
	if err != nil {
		tb.Fatalf("Error acquiring expired lease: %s", err)
	}
	defer provider.Release(ctx, "test", secret3)

	// Can no longer renew the lease with the old secret.
	err = provider.Renew(ctx, "test", secret, t0.Add(20*time.Second))
	if !errors.Is(err, lease.ErrNotHeld) {
		tb.Errorf("got error %v, want ErrNotHeld", err)
	}

	_, err = provider.Acquire(ctx, "test2", t0.Add(20*time.Second))
	if !errors.Is(err, lease.ErrHeld) {
		tb.Errorf("got error %v, want ErrHeld", err)
	}

	err = provider.Release(ctx, "test2", secret2)
	if err != nil {
		tb.Fatalf("Error releasing lease: %s", err)
	}

	mockClock.Add(20 * time.Second) // i.e. t0+35s

	err = provider.Renew(ctx, "test", secret3, t0.Add(50*time.Second))
	if err != nil {
		tb.Fatalf("Error renewing lease: %s", err)
	}

	mockClock.Add(10 * time.Second) // i.e. t0+45s

	_, err = provider.Acquire(ctx, "test", t0.Add(60*time.Second))
	if !errors.Is(err, lease.ErrHeld) {
		tb.Errorf("got error %v, want ErrHeld", err)
	}

	mockClock.Add(10 * time.Second) // i.e. t0+65s

	secret4, err := provider.Acquire(ctx, "test", t0.Add(80*time.Second))
	if err != nil {
		tb.Fatalf("Error acquiring expired lease: %s", err)
	}
	defer provider.Release(ctx, "test", secret4)
}
