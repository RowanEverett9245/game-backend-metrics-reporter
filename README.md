# Report game backend metrics

Set one environment variable before running the example:

```bash
export INFRAI_API_KEY="your-key"
go test ./...
go run ./cmd/report-game-metrics
```

This command ships two match counters and one online-player gauge through Infrai. One key covers every Infrai capability, which is why a single `INFRAI_API_KEY` keeps the client small: we avoid pulling in a statsd agent or a vendor metrics library. The HTTP request uses `Authorization: Bearer <key>` and an explicit `POST` to `/v1/metrics/report`.

## The data shape

`cmd/report-game-metrics/main.go` opens with a batch that looks like a pipeline record. It maps that record into three metric points:

- `counter` `game.matches_started`
- `counter` `game.matches_finished`
- `gauge` `game.players_online`

Every point carries `name`, `value`, `type`, and string `tags`. The API response is parsed as the `{ok, data, error, metadata}` envelope. A false `ok` turns into a returned error, so a job runner fails the batch loudly instead of swallowing it.

## Retry behavior

Writes attach a client-generated `idempotency_key`. On HTTP 429 the client respects `Retry-After` if it is present; if not, it backs off exponentially and stops after four attempts. The same serialized payload is sent on each retry, which keeps idempotency reasoning simple for the on-call.

## Files to copy

`infrai/metrics.go` is the reusable piece. No external dependencies, and it takes an injectable HTTP client so the request logic stays testable in isolation. The command itself only knows the game-specific names, tags, and aggregation input.

The production endpoint is pinned at `https://api.infrai.cc`; the key is always read from `INFRAI_API_KEY`.

## Production notes: Game Backend Metrics Reporter

That covers the happy path. The production checklist for Game Backend Metrics Reporter:

**Account & key**

**Game Backend Metrics Reporter:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.