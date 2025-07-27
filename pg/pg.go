// Package pg implements [lease.Provider] in terms of a PostgresQL database.
package pg

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/bobg/errors"

	"github.com/bobg/lease"
)

// Provider is a lease.Provider implemented in terms of a PostgresQL database.
type Provider struct {
	lease.Clock

	table string // name of the table that stores leases
	db    *sql.DB
	done  chan struct{}
}

var _ lease.Provider = &Provider{}

// New creates a new PostgresQL lease provider.
// Leases are stored in a table with the given name.
// The table is created if it does not already exist.
func New(ctx context.Context, db *sql.DB, table string, opts ...Option) (*Provider, error) {
	const qfmt = `CREATE TABLE IF NOT EXISTS %s (
		name TEXT NOT NULL PRIMARY KEY,
		secret TEXT NOT NULL,
		exp_secs BIGINT NOT NULL
	)`
	q := fmt.Sprintf(qfmt, table)

	if _, err := db.ExecContext(ctx, q); err != nil {
		return nil, errors.Wrapf(err, "creating table %s", table)
	}

	ch := make(chan struct{})

	p := &Provider{
		Clock: lease.DefaultClock{},
		table: table,
		db:    db,
		done:  ch,
	}

	for _, opt := range opts {
		opt(p)
	}

	go func() {
		for {
			select {
			case <-ch:
				return

			case <-p.After(5 * time.Minute):
				const qfmt = `DELETE FROM %s WHERE exp_secs < EXTRACT(EPOCH FROM NOW())`
				q := fmt.Sprintf(qfmt, table)
				_, _ = db.ExecContext(ctx, q)

			}
		}
	}()

	return p, nil
}

// Option is the type of an option that can be passed to [New].
type Option func(*Provider)

// WithClock is an [Option] that sets the clock used by the provider.
func WithClock(c lease.Clock) Option {
	return func(p *Provider) {
		p.Clock = c
	}
}

// Close releases resources held by the provider.
// However, it does _not_ close the underlying database connection.
func (p *Provider) Close() {
	if p.done != nil {
		close(p.done)
		p.done = nil // make this call idempotent
	}
}

func (p *Provider) Acquire(ctx context.Context, name string, exp time.Time) (string, error) {
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(exp) {
		exp = deadline
	}
	deadlineSecs := exp.Unix()

	var secretBytes [16]byte
	if _, err := rand.Read(secretBytes[:]); err != nil {
		return "", errors.Wrap(err, "generating secret")
	}
	secret := hex.EncodeToString(secretBytes[:])

	const qfmt = `
		INSERT INTO %s (name, secret, exp_secs) VALUES ($1, $2, $3)
			ON CONFLICT (name) DO UPDATE SET secret = $2, exp_secs = $3
				WHERE leases.exp_secs < EXTRACT(EPOCH FROM NOW())`
	q := fmt.Sprintf(qfmt, p.table)

	res, err := p.db.ExecContext(ctx, q, name, secret, deadlineSecs)
	if err != nil {
		return "", errors.Wrapf(err, "acquiring lease", name)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return "", errors.Wrapf(err, "counting affected rows", name)
	}
	if aff == 0 {
		return "", lease.ErrHeld
	}

	return secret, nil
}

func (p *Provider) Renew(ctx context.Context, name, secret string, exp time.Time) error {
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(exp) {
		exp = deadline
	}

	var (
		expSecs = exp.Unix()
		nowSecs = p.Now().Unix()
	)

	const qfmt = `UPDATE %s SET exp_secs = $1 WHERE name = $2 AND secret = $3 AND exp_secs > $4`
	q := fmt.Sprintf(qfmt, p.table)

	res, err := p.db.ExecContext(ctx, q, expSecs, name, secret, nowSecs)
	if err != nil {
		return errors.Wrapf(err, "renewing lease", name)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "counting affected rows", name)
	}
	if aff == 0 {
		return lease.ErrNotHeld
	}

	return nil
}

func (p *Provider) Release(ctx context.Context, name, secret string) error {
	const qfmt = `DELETE FROM %s WHERE name = $1 AND secret = $2`
	q := fmt.Sprintf(qfmt, p.table)

	res, err := p.db.ExecContext(ctx, q, name, secret)
	if err != nil {
		return errors.Wrapf(err, "releasing lease", name)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "counting affected rows", name)
	}
	if aff == 0 {
		return lease.ErrNotHeld
	}

	return nil
}
