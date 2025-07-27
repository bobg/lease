package lease_test

import (
	"context"
	"testing"

	"github.com/bobg/lease"
	"github.com/bobg/lease/mem"
	"github.com/bobg/lease/testutil"
)

func factory(clock lease.Clock) (lease.Provider, error) {
	p := mem.New()
	p.Clock = clock
	return p, nil
}

func TestLeader(t *testing.T) {
	testutil.Leader(context.Background(), t, factory)
}
