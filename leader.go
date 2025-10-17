package lease

import (
	"context"
	"time"

	"github.com/bobg/errors"
	"github.com/bobg/retry"
)

// Leader permits running a function after winning a leader election.
// See [Leader.Run].
type Leader struct {
	Name   string        // name for the lease to acquire
	Dur    time.Duration // how long the lease should be valid for
	Retry  time.Duration // how often to retry acquiring the lease
	Jitter time.Duration // plus or minus this much jitter on the retry delay
	Renew  time.Duration // how often to renew the lease after acquiring it; should be less than Dur
}

// Run runs a function after winning a leader election.
//
// The election happens by trying to acquire a lease from the given [Provider]
// using the Name field of l.
// If the lease is already held by another caller,
// Run will retry indefinitely at l.Retry intervals (plus or minus up to l.Jitter),
// until the context is canceled or the lease is acquired.
//
// Once the lease is acquired, Run will renew it periodically at l.Renew intervals.
//
// The provided function f is run with a context that is canceled if the lease cannot be renewed.
// If this happens, [context.Cause] will return a [RenewError] wrapping the error from [Provider.Renew].
//
// The boolean result from Run indicates whether f was ever called.
// If f was called and returned an error,
// that error is wrapped in a [CallbackError] and returned by Run.
// (That that may be a [RenewError] wrapping yet another error,
// if f encountered it and chose to return it.)
func (l Leader) Run(ctx context.Context, p Provider, f func(context.Context) error) (bool, error) {
	tr := retry.Tryer{
		Max:         -1,       // retry indefinitely
		Delay:       l.Retry,  // this often
		Jitter:      l.Jitter, // plus or minus (up to) this much
		IsRetryable: func(e error) bool { return errors.Is(e, ErrHeld) },
		After:       p.After,
	}

	var secret string

	err := tr.Try(ctx, func(int) error {
		var err error
		secret, err = p.Acquire(ctx, l.Name, p.Now().Add(l.Dur))
		return err
	})
	if err != nil {
		return false, errors.Wrap(err, "acquiring lease")
	}

	// Lease is acquired.

	defer p.Release(ctx, l.Name, secret)

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	// Renew the lease periodically.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case <-p.After(l.Renew):
				if err := p.Renew(ctx, l.Name, secret, p.Now().Add(l.Dur)); err != nil {
					cancel(RenewError{Err: err})
					return
				}
			}
		}
	}()

	err = f(ctx)
	if err != nil {
		err = CallbackError{Err: err}
	}
	return true, err
}

// RenewError is a wrapper for the error from [Provider.Renew]
// when the lease in [Leader.Run] cannot be renewed.
type RenewError struct {
	Err error
}

func (e RenewError) Error() string { return "renewing lease: " + e.Err.Error() }
func (e RenewError) Unwrap() error { return e.Err }

// CallbackError is a wrapper for the error returned by the callback function to [Leader.Run].
type CallbackError struct {
	Err error
}

func (e CallbackError) Error() string { return "leader callback: " + e.Err.Error() }
func (e CallbackError) Unwrap() error { return e.Err }
