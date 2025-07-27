// Package mem implements [lease.Provider] in memory.
package mem

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/bobg/errors"

	"github.com/bobg/lease"
)

type (
	// Provider is a lease.Provider implemented in memory.
	Provider struct {
		lease.Clock

		mu     sync.Mutex
		leases map[string]leasePair
	}

	leasePair struct {
		secret string
		exp    time.Time
	}
)

var _ lease.Provider = &Provider{}

// New creates a new in-memory lease provider.
func New() *Provider {
	return &Provider{
		Clock:  lease.DefaultClock{},
		leases: make(map[string]leasePair),
	}
}

func (p *Provider) Acquire(ctx context.Context, name string, exp time.Time) (string, error) {
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(exp) {
		exp = deadline
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	pair, ok := p.leases[name]
	if ok && pair.exp.After(p.Now()) {
		return "", lease.ErrHeld
	}

	var secretBytes [16]byte
	if _, err := rand.Read(secretBytes[:]); err != nil {
		return "", errors.Wrap(err, "generating secret")
	}
	secret := hex.EncodeToString(secretBytes[:])

	p.leases[name] = leasePair{
		secret: secret,
		exp:    exp,
	}

	return secret, nil
}

func (p *Provider) Renew(ctx context.Context, name, secret string, exp time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	pair, isHeld := p.isHeld(name, secret)
	if !isHeld {
		return lease.ErrNotHeld
	}

	if deadline, ok := ctx.Deadline(); ok && deadline.Before(exp) {
		exp = deadline
	}

	pair.exp = exp
	p.leases[name] = pair

	return nil
}

func (p *Provider) Release(_ context.Context, name, secret string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, isHeld := p.isHeld(name, secret); !isHeld {
		return lease.ErrNotHeld
	}

	delete(p.leases, name)

	return nil
}

// Precondition: the caller must hold the mutex.
func (p *Provider) isHeld(name, secret string) (leasePair, bool) {
	pair, ok := p.leases[name]
	return pair, ok && pair.secret == secret && pair.exp.After(p.Now())
}
