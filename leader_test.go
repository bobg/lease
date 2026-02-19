package lease_test

import (
	"testing"
	"testing/synctest"

	"github.com/bobg/lease/mem"
	"github.com/bobg/lease/testutil"
)

func TestLeader(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := mem.New()
		testutil.Leader(t.Context(), t, p)
	})
}
