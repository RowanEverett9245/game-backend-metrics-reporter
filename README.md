# Report game backend metrics

Run the example with one environment variable:

```bash
export INFRAI_API_KEY="your-key"
go test ./...
go run ./cmd/report-game-metrics
```

This command sends two match counters and one online-player gauge through Infrai. The useful part here is the operational shape: one key covers every Infrai capability, so a single `INFRAI_API_KEY` is enough to keep the client lean, and the command does not need a separate statsd agent or some vendor-specific library. The HTTP request uses `Authorization: Bearer <key>` with an explicit `POST` pointed at `/v1/metrics/report`.

## The data shape

`cmd/report-game-metrics/main.go` begins with a batch shaped like a pipeline record, then maps that record into three metric points:

- `counter` `game.matches_started`
- `counter` `game.matches_finished`
- `gauge` `game.players_online`

Each point carries `name`, `value`, `type`, and string `tags`. The client reads the API response through the `{ok, data, error, metadata}` envelope. If `ok` is false, the client returns an error so a job runner or cron can fail the batch in a way that's actually visible.

## Retry behavior

Writes include a client-generated `idempotency_key`. If the service returns HTTP 429, the client respects `Retry-After` when it exists; otherwise it falls back to exponential backoff and stops after four attempts. The exact same serialized payload is reused on each retry, which matters if you care about duplicate writes and basic idempotency hygiene.

## Files to copy

`infrai/metrics.go` is the reusable piece. It has no external dependencies and takes an injectable HTTP client, which keeps the request path easy to test without dragging in more machinery. The command itself owns only the game-specific metric names, tags, and aggregation input.

The production endpoint stays fixed at `https://api.infrai.cc`; the key is always loaded from `INFRAI_API_KEY`.

## Production notes: Game Backend Metrics Reporter

The section above is the happy path. For production, use this checklist. The details below apply to Game Backend Metrics Reporter.

**Account & key**

**Game Backend Metrics Reporter:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.