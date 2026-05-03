# Neurofeed — workflows and procedures

[Index of all docs](README.md). Short guide for humans and coding agents. Coding standards live in [RULES.md](RULES.md). Phase delivery status: [SUMMARY.md](../SUMMARY.md) (update when each phase ships).

Step-by-step local run (prerequisites, `.env`, getting `TELEGRAM_CHAT_ID`): [HOW_TO_RUN.md](HOW_TO_RUN.md).

## Local run (phase 0+)

```bash
make run
# or
go run ./cmd/neurofeed
```

Build binary:

```bash
make build
./bin/neurofeed
```

## Quality commands

```bash
make fmt    # go fmt ./...
make vet    # go vet ./...
make test   # go test ./... -count=1
make all    # fmt, vet, test, build
```

## Environment

See [.env.example](../.env.example). Example:

```bash
export TELEGRAM_BOT_TOKEN="your-token"
export TELEGRAM_CHAT_ID="your-chat-id"
export LLM_PROVIDER="openai"
export LLM_MODEL="gpt-4o-mini"
export LLM_BASE_URL="https://api.openai.com/v1"
export LLM_API_KEY="your-key"
export RSS_FEED_URL="https://example.com/feed.xml"
# Optional: RSS_FEED_TIER=primary|expert|news|community (default news)

# Multiple feeds (JSON; takes precedence over RSS_FEED_URL):
# export NEUROFEED_RSS_FEEDS='[{"url":"https://a.example/feed","tier":"primary"},{"url":"https://b.example/atom","tier":"news"}]'
```

Optional:

- `NEUROFEED_HTTP_TIMEOUT` — Go duration string (e.g. `45s`) for default HTTP client timeout baseline.
- `NEUROFEED_HTTP_TIMEOUT_SECONDS` — integer seconds; if set after the duration env is parsed, it overrides the timeout (see `internal/config`).
- `NEUROFEED_LLM_TIMEOUT` — Go duration for the **LLM-only** HTTP client (default `60s`; used by `-llm-smoke` and later the real summarizer).
- `NEUROFEED_RSS_ITEMS_PER_FEED` — integer **N** newest items **per RSS URL** after parse (sorted by `pubDate` / `updated`; default **2**). Set to **0** to disable the cap and keep every item each feed returns.

**OpenAI I/O check (phase 3.1):** with `LLM_API_KEY` set (and optional `LLM_BASE_URL`, `LLM_MODEL`, `LLM_PROVIDER=openai`), run `make llm-smoke` or `go run ./cmd/neurofeed -llm-smoke` — one `chat/completions` call, no RSS/Telegram required.

**Full digest run:** same `LLM_*` plus RSS/Telegram — `make run` uses **`DigestSummarizer`** when the provider is OpenAI (or unset). Tune **`NEUROFEED_LLM_MAX_ARTICLES`** and **`NEUROFEED_LLM_MAX_OUTPUT_TOKENS`** if needed (with multiple **`subject`** values, articles are **sampled round-robin by subject** up to that cap so every topic reaches the model). Telegram messages use **HTML** parse mode (bold headers, clickable links, escaped text).

## Adding an RSS source

1. Add the feed URL to configuration (`RSS_FEED_URL` / `RSS_FEED_TIER`, or an entry in `NEUROFEED_RSS_FEEDS` JSON). Optionally set **`subject`** on each JSON object (or **`RSS_FEED_SUBJECT`** for the single-feed path): feeds with at least one non-empty subject get **sectioned digests** (`📌` headers; LLM returns `sections` with **`picks`** per subject). **Every distinct `subject` from the feed list appears in each digest** (in first-configured order), even when that feed produced no items—empty sections show a short placeholder. Feeds with no subject still work; items without a subject bucket as **`Geral`** when any other feed has a subject.
2. Confirm fetch timeout and `User-Agent` in the ingest package.
3. Run `make test` and a manual `make run` in a safe environment.

## Scheduler options (later)

- **OS cron / systemd timer**: invoke `bin/neurofeed` on a schedule; simplest ops model.
- **In-process gocron**: embed scheduling in the binary; fewer moving parts on a single host.

Document the chosen approach here once the project picks one.

## Recipients and fixed subjects (phase 5)

Product detail lives in [neurofeed.md](neurofeed.md) (Fase 5 + *Temas e destinatários* / *Fontes por tema*). Summary for implementers:

- **Telegram users do not pick topics** in the bot. They only **receive** formatted digests; **section labels (subjects)** and **which RSS URLs** apply come entirely from **operator config** (versioned file or code), per recipient when multi-`chat_id` exists.
- **Feeds per subject** are a **fixed curated list** (URLs + tiers) — **no** LLM for feed discovery. The pipeline resolves each recipient’s feed set from their assigned subjects, dedupes by URL.
- **Changing subjects or feeds** is an **operator change** (redeploy / edit config), not an in-chat flow.

Until multi-recipient config exists, global `RSS_FEED_URL` / `NEUROFEED_RSS_FEEDS` remain the MVP way to configure feeds (see *Adding an RSS source* above).

## Prompt templates (phase 3.2)

Phase **3.1** adds the OpenAI HTTP client and env-backed settings; **3.2** adds digest prompts from the product brief alongside that client under `internal/ai/`. To A/B prompts, use configuration to select template name or file path.

## Rotating secrets

1. Revoke old token at the provider.
2. Update the environment where the binary runs (no code change).
3. Restart the job or service.

## Definition of done (by phase)

Aligns with [neurofeed.md](neurofeed.md) and [the development plan](../.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md).

- **Phase 0**: `make all` passes; binary runs; stubs respect `context`.
- **Phase 1**: one RSS → Telegram message path works with env config.
- **Later phases**: each completes the checklist rows in the plan without breaking prior phases.

## Cursor skills (optional)

If you add repo-specific Agent Skills under `.cursor/skills/`, list them here with path and one-line purpose.
