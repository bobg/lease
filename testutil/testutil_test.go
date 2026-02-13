package testutil

import (
"context"
"testing"

"github.com/bobg/lease/mem"
)

// These tests already appear elsewhere in this library.
// Duplicating them here solves a problem in how test coverage is measured.

func TestLeader(t *testing.T) {
p := mem.New()
Leader(context.Background(), t, p)
}

func TestProvider(t *testing.T) {
p := mem.New()
Provider(context.Background(), t, p)
}
