package testutil

import (
	"testing"
	"testing/synctest"

	"github.com/bobg/lease/mem"
)

// These tests already appear elsewhere in this library.
// Duplicating them here solves a problem in how test coverage is measured.

func TestLeader(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := mem.New()
		Leader(t.Context(), t, p)
	})
}

func TestProvider(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := mem.New()
		Provider(t.Context(), t, p)
	})
}
