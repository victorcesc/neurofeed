# Instructions for AI coding agents — Neurofeed

## Before writing code

1. Read [RULES.md](RULES.md) for layout, errors, context, HTTP, and testing expectations.
2. Read [SKILLS.md](SKILLS.md) for run commands and env vars.
3. Read [neurofeed.md](neurofeed.md) for product behavior and prompts.

## Language and toolchain

- Primary language: **Go** (module minimum version in [go.mod](go.mod); use a supported stable Go release locally).
- Commands: `make fmt`, `make vet`, `make test`, `make run` / `make build`.

## Scope and style

- Prefer **minimal diffs** that match existing patterns in the same package.
- Do not add documentation files unless the user or plan explicitly asks.
- Do not commit secrets or paste real tokens into chat, logs, or tests.

## Architecture boundaries

- **ingest** → **domain** → **ai** → **notify**, orchestrated by **pipeline** (see plan diagram in `.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md`).
- Introduce small interfaces at package boundaries for testability (`FeedFetcher`, `Summarizer`, `Notifier`).
- All external I/O accepts `context.Context`.

## When to add packages

- Extend an existing `internal/` package when the change fits its responsibility.
- Add a new `internal/<name>/` directory when a new external system or clear sub-boundary appears (e.g. a second notifier).

## Security

- Treat `TELEGRAM_BOT_TOKEN`, `OPENAI_API_KEY`, and feed URLs as sensitive.
- Use placeholders in examples (`<token>`), never real credentials.
