package mem

import (
	"context"
	"testing"

	"github.com/bobg/lease"
	"github.com/bobg/lease/testutil"
)

func factory(clock lease.Clock) (lease.Provider, error) {
	p := New()
	p.Clock = clock
	return p, nil
}

func TestProvider(t *testing.T) {
	testutil.Provider(context.Background(), t, factory)
}
