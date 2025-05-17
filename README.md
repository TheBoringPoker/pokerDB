# pokerDB

This repository contains a small poker toolkit written in Go. The project focuses on describing poker games and now includes a basic data layer using GORM to persist games and actions across multiple SQL databases.

## Building and Testing

This project uses Go modules. To run the unit tests:

```bash
go test ./...
```

The tests require network access to download dependencies (GORM and database drivers). In completely offline environments they will fail during module download.

## Database Integration

The `storage` package provides a simple wrapper around GORM with support for MySQL, PostgreSQL and SQLite. Configure the `Dialect` and `DSN` fields in `storage.Config` to connect to the desired backend. SQLite can be used for local testing without any external services:

```go
cfg := storage.Config{Dialect: storage.DialectSQLite, DSN: "file:test.db?cache=shared&mode=memory"}
db, err := storage.NewDB(cfg)
```

Tables are created automatically using `AutoMigrate`. UUID primary keys are generated in Go so that SQLite works out of the box without additional extensions.

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

`Game.ActionStrings()` formats these entries into human readable lines.

## Project Structure

- `cmd/` – example main program
- `pkg/models/` – data models used by the application
- `pkg/storage/` – database connection helpers
- `pkg/rules/` – poker evaluation and game rules
- `pkg/utils/` – utility helpers


