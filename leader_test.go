package lease_test

import (
"context"
"testing"

"github.com/bobg/lease/mem"
"github.com/bobg/lease/testutil"
)

func TestLeader(t *testing.T) {
p := mem.New()
testutil.Leader(context.Background(), t, p)
}
