# Neurofeed

CLI tool that pulls **RSS/Atom** sources, **deduplicates** items, optionally builds an **LLM digest**, and sends **HTML-formatted** summaries to **Telegram**. Written in **Go**; pipeline layout is `ingest` → `domain` → `ai` → `notify`, orchestrated by `internal/pipeline`.

## Features

- **Multiple feeds** via `NEUROFEED_RSS_FEEDS` (JSON) or a single `RSS_FEED_URL`, with optional **source tier** per feed (`primary`, `expert`, `news`, `community`).
- **Sectioned digests** when feeds declare a **`subject`** (or `RSS_FEED_SUBJECT` for the single-feed path); items without a subject bucket as `Geral`.
- **Per-feed item cap** (`NEUROFEED_RSS_ITEMS_PER_FEED`, default 2 newest per URL) and **title dedup** after merge.
- **OpenAI-compatible** chat completion for JSON digest output; **headline fallback** when no LLM key or non-OpenAI provider.
- **Telegram** delivery with `parse_mode: HTML` (bold titles, links, two summary lines per pick when the LLM path is active).
- **`-llm-smoke`** for a minimal completion test without running the full RSS → Telegram pipeline.

Roadmap and what is shipped per phase: **[SUMMARY.md](SUMMARY.md)**. Product rules and prompts: **[docs/neurofeed.md](docs/neurofeed.md)**.

## Requirements

- **Go** ≥ version in [go.mod](go.mod) (currently 1.23).
- **GNU Make** (optional): [Makefile](Makefile) wraps `fmt`, `vet`, `test`, `build`, `run`, `llm-smoke`.

## Quick start

```bash
git clone https://github.com/victorcesc/neurofeed.git
cd neurofeed
go mod download
cp .env.example .env
# Edit .env: at minimum TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID, and feed URL(s) or NEUROFEED_RSS_FEEDS.
set -a && source .env && set +a   # bash/zsh; adjust for your shell
make run
```

A successful pipeline run logs `step=pipeline_run_ok`. Use **Ctrl+C** or SIGTERM for graceful shutdown.

**Test the LLM only** (no RSS/Telegram):

```bash
set -a && source .env && set +a
make llm-smoke
# equivalent: go run ./cmd/neurofeed -llm-smoke
```

Full variable reference, Telegram `chat_id` notes, and troubleshooting: **[docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md)**. Compact env and workflow index: **[docs/SKILLS.md](docs/SKILLS.md)**.

## Configuration

Copy [.env.example](.env.example) to `.env` and **never commit** real tokens or feed secrets. Important variables:

| Area | Variables (see `.env.example` for full list) |
|------|-----------------------------------------------|
| Telegram | `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` |
| Feeds | `RSS_FEED_URL` / `RSS_FEED_TIER` / `RSS_FEED_SUBJECT` or `NEUROFEED_RSS_FEEDS` |
| LLM | `LLM_API_KEY`, `LLM_BASE_URL`, `LLM_MODEL`, `LLM_PROVIDER`, `NEUROFEED_LLM_TIMEOUT`, `NEUROFEED_LLM_MAX_ARTICLES`, `NEUROFEED_LLM_MAX_OUTPUT_TOKENS` |
| HTTP | `NEUROFEED_HTTP_TIMEOUT`, `NEUROFEED_HTTP_TIMEOUT_SECONDS` |
| Ingest | `NEUROFEED_RSS_ITEMS_PER_FEED` |

## Development

```bash
make fmt      # go fmt ./...
make vet      # go vet ./...
make test     # go test ./... -count=1
make build    # outputs bin/neurofeed
make all      # fmt, vet, test, build
```

Engineering conventions: **[docs/RULES.md](docs/RULES.md)**. Architecture diagram and phased plan: **[.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md](.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md)**.

## Repository layout

| Path | Role |
|------|------|
| `cmd/neurofeed/` | CLI entrypoint, flags (`-llm-smoke`), wiring |
| `internal/config/` | Environment loading and validation |
| `internal/ingest/` | RSS fetch (`gofeed`) |
| `internal/domain/` | Article model, dedup, balancing across subjects |
| `internal/ai/` | OpenAI client, digest JSON parsing, summarizers |
| `internal/notify/` | Telegram notifier |
| `internal/pipeline/` | End-to-end orchestration |
| `docs/` | Product spec, runbooks, conventions |

## Documentation

| Document | Purpose |
|----------|---------|
| [SUMMARY.md](SUMMARY.md) | Phase delivery status; update when a phase ships |
| [docs/neurofeed.md](docs/neurofeed.md) | Product behavior, operator-hardcoded feeds/subjects, Telegram receive-only |
| [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md) | Local setup, env vars, Telegram |
| [docs/SKILLS.md](docs/SKILLS.md) | Commands, env index, checklists |
| [docs/RULES.md](docs/RULES.md) | Go and repo conventions |

**AI coding agents:** start with [AGENTS.md](AGENTS.md).
