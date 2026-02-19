package mem

import (
	"testing"
	"testing/synctest"

	"github.com/bobg/lease/testutil"
)

func TestProvider(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := New()
		testutil.Provider(t.Context(), t, p)
	})
}
