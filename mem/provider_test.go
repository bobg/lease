//go:build goexperiment.synctest

package mem

import (
"context"
"testing"

"github.com/bobg/lease/testutil"
)

func TestProvider(t *testing.T) {
p := New()
testutil.Provider(context.Background(), t, p)
}
