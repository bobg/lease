package pg

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
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

func TestQueryWithExpSecs(t *testing.T) {
	mockClock := clock.NewMock()
	mockClock.Set(time.Date(1977, 8, 5, 0, 0, 0, 0, time.UTC))

	cases := []struct {
		clock     lease.Clock
		qfmt      string
		qargs     []any
		wantQuery string
		wantQargs []any
	}{{
		clock:     lease.DefaultClock{},
		qfmt:      `SELECT * FROM %s WHERE exp_secs < %s`,
		wantQuery: `SELECT * FROM table WHERE exp_secs < EXTRACT(EPOCH FROM NOW())`,
	}, {
		clock:     mockClock,
		qfmt:      `SELECT * FROM %s WHERE exp_secs < %s`,
		wantQuery: `SELECT * FROM table WHERE exp_secs < $1`,
		wantQargs: []any{mockClock.Now().Unix()},
	}, {
		clock:     lease.DefaultClock{},
		qfmt:      `UPDATE %s SET secret = $1, exp_secs = $2 WHERE name = $3 AND exp_secs < %s`,
		qargs:     []any{"foo", 1, "bar"},
		wantQuery: `UPDATE table SET secret = $1, exp_secs = $2 WHERE name = $3 AND exp_secs < EXTRACT(EPOCH FROM NOW())`,
		wantQargs: []any{"foo", 1, "bar"},
	}, {
		clock:     mockClock,
		qfmt:      `UPDATE %s SET secret = $1, exp_secs = $2 WHERE name = $3 AND exp_secs < %s`,
		qargs:     []any{"foo", 1, "bar"},
		wantQuery: `UPDATE table SET secret = $1, exp_secs = $2 WHERE name = $3 AND exp_secs < $4`,
		wantQargs: []any{"foo", 1, "bar", mockClock.Now().Unix()},
	}}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case_%02d", i+1), func(t *testing.T) {
			p := &Provider{
				Clock: tc.clock,
				table: "table",
			}

			gotQuery, gotQargs := p.queryWithExpSecs(tc.qfmt, tc.qargs)
			if gotQuery != tc.wantQuery {
				t.Errorf("got query %q, want %q", gotQuery, tc.wantQuery)
			}
			if !slices.Equal(gotQargs, tc.wantQargs) {
				t.Errorf("got args %v; want %v", gotQargs, tc.wantQargs)
			}
		})
	}
}
