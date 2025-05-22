# pokerDB

This project was created with the help of the **OpenAI Codex** system under the direction of the author. It is an experimental coding agent project that provides a foundation data model for poker games. No functionality is guaranteed.

The repository contains a small poker toolkit written in Go. The project focuses on describing poker games and now includes a basic data layer using GORM to persist games and actions across multiple SQL databases.

## Building and Testing

This project uses Go modules. To run the unit tests:

```bash
go test ./...
go test -race ./...
```

All tests run purely in-memory and no external database server is required. The
integration test in `pkg/storage` uses SQLite and is skipped by default because
the SQLite driver cannot auto migrate UUID defaults.

The tests require network access to download dependencies (GORM and database drivers). In completely offline environments they will fail during module download.
SQLite integration tests rely on CGO for the `go-sqlite3` driver. When `CGO_ENABLED=0` the tests will be automatically skipped.

The example command line program can be built on any platform supported by Go. The toolkit has been used on Ubuntu and macOS with Go 1.20:

```bash
go build ./cmd
```

You may also run it directly:

```bash
go run ./cmd
```

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

Package `pkg/models` contains a multi-threaded test (`TestGameConcurrentInvalidOps`) that stresses these methods.

## Compact Action Log

Each `Game` stores a slice of encoded strings describing every event. The first
entry records the game options (blinds, ante, whether run-it-twice or straddle
are allowed) and the last entry captures final ledger balances. Intermediate
entries encode player actions such as raise, fold, check, all-in, straddle,
buy-ins and run-it-twice selections. Example entries:

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

## Buy-In Handling

Games define `MinBuyIn` and `MaxBuyIn` limits. Buy-ins may occur at any
time using `Game.BuyIn`. Each buy-in is logged with a `B` entry.
Starting the game still requires one buy-in per seat and any amounts
outside the allowed range cause an error.

Players can join or leave a running game with `Game.Join` and
`Game.Quit`. These actions are recorded in the log using the `J` and `Q`
codes. If the last player quits, the game is automatically ended.

Seat selections for the next game can be made with `Game.ChooseSeat`.
Seats are numbered 1-9 and a seat may only be selected by one player.
The chosen seat number is logged with the `H` action code for historic
tracking.

## Action Validation

The `validate` package checks recorded actions for consistency. In
addition the `Game` type now validates chip counts as actions are
recorded so a bet or call cannot exceed the acting player's stack.
`validate.Validate` can still be used to audit a completed game. Pass a
map of starting chip counts keyed by the truncated player IDs found in
the action log.

## Project Structure

- `cmd/` – example main program
- `pkg/models/` – data models used by the application
- `pkg/storage/` – database connection helpers
- `pkg/rules/` – poker evaluation and game rules
- `pkg/utils/` – utility helpers
- `pkg/rules/validate/` – action log validation helpers


## Disclaimer

This repository is provided as an educational experiment. It was generated using OpenAI Codex and assembled by the author. The software is offered as-is with no guarantee of correctness or fitness for any purpose.


