// Package lease defines a [Provider] that supplies leases, which are timed mutual-exclusion locks.
// It also defines [Leader], a simple leader-election mechanism.
package lease

import (
	"context"
	"errors"
	"time"
)

// Provider is the type of a lease provider.
type Provider interface {
	Clock

	// Acquire acquires a lease if available and returns a secret, needed for Renew and Release.
	// The lease expires at the given time,
	// or at the deadline of the provided context (if it has one), whichever is earlier.
	// Acquire does not wait; if the lease is already held by another caller, it returns [ErrHeld].
	Acquire(context.Context, string, time.Time) (string, error)

	// Renew extends the lease for the given name, using the provided secret.
	// Its expiration time is reset to the given time
	// or the deadline of the provided context (if it has one), whichever is earlier.
	// If the lease is not held by the caller, it returns [ErrNotHeld].
	Renew(ctx context.Context, name, secret string, exp time.Time) error

	// Release releases the lease for the given name, using the provided secret.
	// If the lease is not held by the caller, it returns [ErrNotHeld].
	Release(ctx context.Context, name, secret string) error
}

var (
	// ErrHeld is the error returned by [Provider.Acquire] when the lease is already held by another caller.
	ErrHeld = errors.New("lease already held by another caller")

	// ErrNotHeld is the error returned by [Provider.Renew] or [Provider.Release] when the lease is not held by the caller.
	ErrNotHeld = errors.New("lease not held by caller")
)
