# pokerDB

This repository contains a small poker toolkit written in Go. The project focuses on describing poker games and now includes a basic data layer using GORM to persist games and actions across multiple SQL databases.

## Building and Testing

This project uses Go modules. To run the unit tests:

```bash
go test ./...
```

All tests run purely in-memory and no external database server is required. The
integration test in `pkg/storage` uses SQLite and is skipped by default because
the SQLite driver cannot auto migrate UUID defaults.

The tests require network access to download dependencies (GORM and database drivers). In completely offline environments they will fail during module download.
SQLite integration tests rely on CGO for the `go-sqlite3` driver. When `CGO_ENABLED=0` the tests will be automatically skipped.

## Database Integration

The `storage` package provides a simple wrapper around GORM with support for MySQL, PostgreSQL and SQLite. Configure the `Dialect` and `DSN` fields in `storage.Config` to connect to the desired backend. SQLite can be used for local testing:

```go
cfg := storage.Config{Dialect: storage.DialectSQLite, DSN: "file:test.db?cache=shared&mode=memory"}
db, err := storage.NewDB(cfg)
```

Tables are created automatically using `AutoMigrate`.

## Concurrency

`Game` objects may be used by multiple goroutines. The struct now embeds a
mutex so operations like `Start`, `Deal`, `AddAction` and `ActionStrings` are
safe for concurrent use.

## Compact Action Log

Each `Game` stores a slice of encoded strings describing every event. The first
entry records the game options (blinds, ante, whether run-it-twice or straddle
are allowed) and the last entry captures final ledger balances. Intermediate
entries encode player actions such as raise, fold, check, all-in, straddle and
run-it-twice selections. Example entries:

```
G:50:100:0:1:0,1692300000         // game start with small blind, big blind,
                                   // ante, run-it-twice allowed, straddle allowed
c0ffee00C0,1692300010              // player c0ffee00 checks
c0ffee01R500,1692300020            // player c0ffee01 raises to 500
E:c0ffee00=1000:c0ffee01=-1000,1692300100 // final ledger
```

`Game.ActionStrings()` formats these entries into human readable lines. Example
output:

```
start sb=50 bb=100 ante=0 runTwice=1 straddle=0 at 2023-08-18T15:00:00Z
c0ffee00 check 0 at 2023-08-18T15:00:10Z
c0ffee01 raise 500 at 2023-08-18T15:00:20Z
result [c0ffee00=1000 c0ffee01=-1000] at 2023-08-18T15:01:40Z
```

## Project Structure

- `cmd/` – example main program
- `pkg/models/` – data models used by the application
- `pkg/storage/` – database connection helpers
- `pkg/rules/` – poker evaluation and game rules
- `pkg/utils/` – utility helpers


