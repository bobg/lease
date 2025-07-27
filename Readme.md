# Lease

[![Go Reference](https://pkg.go.dev/badge/github.com/bobg/lease.svg)](https://pkg.go.dev/github.com/bobg/lease)
[![Go Report Card](https://goreportcard.com/badge/github.com/bobg/lease)](https://goreportcard.com/report/github.com/bobg/lease)
[![Tests](https://github.com/bobg/lease/actions/workflows/go.yml/badge.svg)](https://github.com/bobg/lease/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/bobg/lease/badge.svg?branch=main)](https://coveralls.io/github/bobg/lease?branch=main)

This is lease,
a Go library for concurrent, possibly distributed
timed mutual-exclusion locks
and leader elections.

## Usage

In all cases you’ll need a Provider, which provides leases.
This library includes two implementations of Provider
(with more to come):
an in-memory version and a [Postgresql](https://www.postgresql.org/) version.

Acquiring a lease:

```go
secret, err := provider.Acquire(ctx, "leaseName", expirationTime)
if err != nil { ... }
defer provider.Release(ctx, "leaseName", secret)
```

Renewing an already-acquired lease:

```go
err := provider.Renew(ctx, "leaseName", newExpirationTime)
if err != nil { ... }
```

Running a function after winning a leader election:

```go
leader := lease.Leader{
  Name:   "leaseName",
  Dur:    5*time.Minute,
  Retry:  time.Minute,
  Jitter: 5*time.Second,
  Renew:  4*time.Minute,
}
err := leader.Run(ctx, provider, func(ctx context.Context) error {
  fmt.Println("I am the leader")
  ...
})
if err != nil { ... }
```
