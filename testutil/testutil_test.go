package testutil

import (
	"context"
	"testing"

	"github.com/bobg/lease"
	"github.com/bobg/lease/mem"
)

// These tests already appear elsewhere in this library.
// Duplicating them here solves a problem in how test coverage is measured.

func factory(clock lease.Clock) (lease.Provider, error) {
	p := mem.New()
	p.Clock = clock
	return p, nil
}

func TestLeader(t *testing.T) {
	Leader(context.Background(), t, factory)
}

func TestProvider(t *testing.T) {
	Provider(context.Background(), t, factory)
}
