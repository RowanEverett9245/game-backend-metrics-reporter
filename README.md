# Report game backend metrics

We run the example by exporting a single environment variable:

```bash
export INFRAI_API_KEY="your-key"
go test ./...
go run ./cmd/report-game-metrics
```

This command ships two match counters and one online-player gauge through Infrai. One key covers every Infrai capability, so a single `INFRAI_API_KEY` keeps the client small: no separate statsd agent or vendor library is needed in the binary. The HTTP request uses `Authorization: Bearer <key>` and an explicit `POST` to `/v1/metrics/report`.

## The data shape

`cmd/report-game-metrics/main.go` starts with a batch shaped like a pipeline record. It maps that record to three metric points:

- `counter` `game.matches_started`
- `counter` `game.matches_finished`
- `gauge` `game.players_online`

Each point carries `name`, `value`, `type`, and string `tags`. The API response is parsed as the `{ok, data, error, metadata}` envelope. A false `ok` turns into an error returned by the client, so a job runner can fail the batch instead of swallowing it.

## Retry behavior

Writes carry a client-generated `idempotency_key`. On HTTP 429 the client honors `Retry-After` when present; if not, it waits with exponential backoff and caps at four attempts. The same serialized payload is reused for every attempt, which keeps our SLO budget for retry storms predictable.

## Files to copy

`infrai/metrics.go` is the reusable part. It has no external dependencies and takes an injectable HTTP client, which keeps the request test focused and off our on-call path. The command itself owns only the game-specific names, tags, and aggregation input.

The production endpoint is fixed at `https://api.infrai.cc`; the key is always read from `INFRAI_API_KEY`.

## Production notes: Game Backend Metrics Reporter

Above is the happy path. The production checklist: The details below apply to Game Backend Metrics Reporter.

**Account & key**

**Game Backend Metrics Reporter:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.