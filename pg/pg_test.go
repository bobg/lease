package pg

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/bobg/lease"
	"github.com/bobg/lease/testutil"
)

func factory(ctx context.Context, db *sql.DB, table string) func(lease.Clock) (lease.Provider, error) {
	return func(clock lease.Clock) (lease.Provider, error) {
		return New(ctx, db, table, WithClock(clock))
	}
}

func TestProvider(t *testing.T) {
	ctx := context.Background()

	withDB(ctx, t, func(db *sql.DB) {
		testutil.Provider(ctx, t, factory(ctx, db, "leases"))
	})
}

func TestLeader(t *testing.T) {
	ctx := context.Background()

	withDB(ctx, t, func(db *sql.DB) {
		testutil.Leader(ctx, t, factory(ctx, db, "leases"))
	})
}

func withDB(ctx context.Context, t *testing.T, f func(*sql.DB)) {
	var (
		dbname   = os.Getenv("POSTGRES_DB")
		dbuser   = os.Getenv("POSTGRES_USER")
		dbpasswd = os.Getenv("POSTGRES_PASSWORD")
	)
	if dbname == "" || dbuser == "" || dbpasswd == "" {
		t.Skip("POSTGRES_DB, POSTGRES_USER, and POSTGRES_PASSWORD must be set")
	}

	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbuser, dbpasswd, dbname))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	f(db)
}
