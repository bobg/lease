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
		dbhost   = os.Getenv("POSTGRES_HOST")
		dbport   = os.Getenv("POSTGRES_PORT")
		dbname   = os.Getenv("POSTGRES_DB")
		dbuser   = os.Getenv("POSTGRES_USER")
		dbpasswd = os.Getenv("POSTGRES_PASSWORD")
	)

	if dbuser == "" {
		t.Skip("POSTGRES_USER must be set")
	}

	if dbhost == "" {
		dbhost = "localhost"
	}
	if dbport == "" {
		dbport = "5432"
	}
	if dbname == "" {
		dbname = dbuser
	}

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbhost, dbport, dbuser, dbpasswd, dbname))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	f(db)
}
