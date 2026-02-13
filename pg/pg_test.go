package pg

import (
"context"
"database/sql"
"fmt"
"os"
"testing"

_ "github.com/lib/pq"

"github.com/bobg/lease/testutil"
)

func TestProvider(t *testing.T) {
ctx := context.Background()

withDB(ctx, t, func(db *sql.DB) {
provider, err := New(ctx, db, "leases")
if err != nil {
t.Fatal(err)
}
defer provider.Close()
testutil.Provider(ctx, t, provider)
})
}

func TestLeader(t *testing.T) {
ctx := context.Background()

withDB(ctx, t, func(db *sql.DB) {
provider, err := New(ctx, db, "leases")
if err != nil {
t.Fatal(err)
}
defer provider.Close()
testutil.Leader(ctx, t, provider)
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
// This test verifies that queryWithExpSecs always uses the server's clock
// since we no longer have a mock clock option.
p := &Provider{
table: "table",
}

qfmt := `SELECT * FROM %s WHERE exp_secs < %s`
gotQuery, gotQargs := p.queryWithExpSecs(qfmt, nil)
wantQuery := `SELECT * FROM table WHERE exp_secs < EXTRACT(EPOCH FROM NOW())`

if gotQuery != wantQuery {
t.Errorf("got query %q, want %q", gotQuery, wantQuery)
}
if len(gotQargs) != 0 {
t.Errorf("got args %v; want empty", gotQargs)
}

// Test with existing query arguments
qfmt = `UPDATE %s SET secret = $1, exp_secs = $2 WHERE name = $3 AND exp_secs < %s`
qargs := []any{"foo", 1, "bar"}
gotQuery, gotQargs = p.queryWithExpSecs(qfmt, qargs)
wantQuery = `UPDATE table SET secret = $1, exp_secs = $2 WHERE name = $3 AND exp_secs < EXTRACT(EPOCH FROM NOW())`
wantQargs := []any{"foo", 1, "bar"}

if gotQuery != wantQuery {
t.Errorf("got query %q, want %q", gotQuery, wantQuery)
}
if len(gotQargs) != len(wantQargs) {
t.Errorf("got %d args, want %d", len(gotQargs), len(wantQargs))
}
for i := range gotQargs {
if gotQargs[i] != wantQargs[i] {
t.Errorf("arg %d: got %v, want %v", i, gotQargs[i], wantQargs[i])
}
}
}
