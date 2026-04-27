# Neurofeed — engineering rules (Go)

Canonical conventions for this repository. Prefer these over ad-hoc style.

## Layout

- `cmd/neurofeed/` — entrypoint only: wiring, signals, exit codes.
- `internal/config/` — environment-based configuration, validation, defaults.
- `internal/domain/` — core types (`Article`, etc.) and pure logic (dedup keys, scoring).
- `internal/ingest/` — RSS and other fetchers; HTTP clients with timeouts.
- `internal/ai/` — LLM clients and prompt assembly.
- `internal/notify/` — Telegram and other sinks.
- `internal/pipeline/` — orchestration across boundaries.

Add a new package when a boundary or test surface deserves isolation. Do not create `pkg/` for private app code.

## Go version and modules

- `go.mod` declares the minimum Go version for the module; use current stable locally.
- Add third-party modules only with a one-line justification in the commit or PR description.
- Prefer `net/http`, `context`, `log/slog`, and the standard library before pulling dependencies.

## Context and I/O

- Every function that performs or may perform network I/O must accept `context.Context` as its **first parameter**, conventionally named `ctx`.
- Use `context.WithTimeout` for single operations; respect `ctx.Err()` in long loops.
- Do not use blank imports or `_` to discard errors from I/O.

## Errors

- Wrap with `%w` when the caller should inspect causes; use `fmt.Errorf("op: %w", err)`.
- Document sentinel errors on the declaring package when exported.
- Log at the boundary (e.g. `main` or `pipeline`) where the error is final; lower layers return errors.

## HTTP clients

- Construct `http.Client` with explicit `Timeout` or use `http.NewRequestWithContext` with a derived context deadline.
- Set a sensible `User-Agent` for RSS fetches (identify the bot responsibly).
- Retries (phase 7): bounded count, exponential backoff, only for idempotent GET and clear transient status codes.

## Configuration and secrets

- Read secrets from the environment (see [`.env.example`](../.env.example)). Never commit tokens or API keys.
- Validate on startup; fail fast with clear messages.

## Logging

- Use `log/slog` with structured keys (`slog.Info("msg", "key", value)`).
- Log levels: `Info` for high-level pipeline steps, `Debug` for verbose diagnostics, `Error` before exit.

## Telegram and LLM providers

- Respect API rate limits; cap batch sizes and output tokens in configuration.
- Telegram message length: plan splitting or continuation messages when the digest exceeds limits (see product spec in [neurofeed.md](neurofeed.md)).
- Escape or select parse mode carefully for Markdown/HTML.

## Testing

- Table-driven tests for pure functions (scoring, dedup, formatting).
- Use `net/http/httptest` for HTTP-dependent clients.
- Inject `time.Time` or a small clock interface for recency scoring.

## Formatting and static analysis

- `make fmt vet test` before pushing.
- Consider `staticcheck` in CI when the project adds continuous integration.

## Naming

- Variable names must be descriptive and communicate intent clearly in their scope.
- Avoid **single-letter** names and vague placeholders (`v`, `x`, `tmp`, `n` as “some number”) except where Go convention is universal and scope is tiny: `ctx` for `context.Context`, `err` for errors, `i`/`j`/`k` in short index loops, `t` for `*testing.T`, `b` for `*testing.B`, `w`/`r` for `http.ResponseWriter` / `*http.Request`.
- Short **words** are fine when the meaning is obvious (`key`, `ok`, `seen`, `body`)—the goal is to ban *cryptic* one-letter names, not to require multi-word identifiers everywhere.
- For non-idiomatic cases, prefer explicit names such as `articleScores`, `requestTimeout`, or `telegramMessageBody`.
