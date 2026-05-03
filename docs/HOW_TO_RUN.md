# How to run Neurofeed

This guide covers running the Neurofeed CLI locally in detail. For a compact command and env checklist, use [SKILLS.md](SKILLS.md). AI coding agents: start with [AGENTS.md](../AGENTS.md).

Product behavior and prompts are in [neurofeed.md](neurofeed.md). Phased delivery and checklists are in [the Cursor plan](../.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md). Engineering conventions are in [RULES.md](RULES.md).

## Prerequisites

- **Go**: at least the version declared in [go.mod](../go.mod) (currently Go 1.23). Use a supported stable Go release on your machine.
- **GNU Make** (optional but recommended): the [Makefile](../Makefile) wraps common commands.

## Get the code and dependencies

From the repository root:

```bash
cd /path/to/neurofeed
go mod download
```

## Environment variables

Configuration is read from the process environment. See [.env.example](../.env.example) for all variables.

1. Copy the example file (do not commit a real `.env`):

   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your values, then load it before each run, for example:

   ```bash
   set -a && source .env && set +a
   ```

   Or export variables manually:

   ```bash
   export TELEGRAM_BOT_TOKEN="your-token"
   export TELEGRAM_CHAT_ID="your-chat-id"
   export RSS_FEED_URL="https://example.com/feed.xml"
   export LLM_PROVIDER="openai"
   export LLM_MODEL="gpt-4o-mini"
   export LLM_BASE_URL="https://api.openai.com/v1"
   export LLM_API_KEY="your-key"
   ```

**Optional HTTP tuning**

- `NEUROFEED_HTTP_TIMEOUT` — Go duration string (for example `45s`).
- `NEUROFEED_HTTP_TIMEOUT_SECONDS` — positive integer seconds; if set, it overrides the duration-based value after parsing (see `internal/config`).
- `NEUROFEED_LLM_TIMEOUT` — Go duration for OpenAI chat completion requests (default `60s`); see **LLM smoke** below.

## Run the application

**Recommended**

```bash
make run
```

**Equivalent**

```bash
go run ./cmd/neurofeed
```

The process listens for **SIGINT** and **SIGTERM** for graceful shutdown context; stop it with Ctrl+C or `kill` as usual.

### RSS → Telegram (no LLM digest yet)

Required environment variables:

- **Feeds** — either:
  - `RSS_FEED_URL` — one RSS or Atom URL; optional `RSS_FEED_TIER` (`primary`, `expert`, `news`, `community`; default `news`), or
  - `NEUROFEED_RSS_FEEDS` — JSON array of `{"url":"...","tier":"..."}` and optional **`"subject":"…"`** (section label, up to 64 characters). When at least one feed has `subject`, the digest is grouped by those labels (`Geral` buckets items from feeds with no subject). For a single `RSS_FEED_URL`, optional **`RSS_FEED_SUBJECT`** does the same.
- `TELEGRAM_BOT_TOKEN` — from BotFather.
- `TELEGRAM_CHAT_ID` — where `sendMessage` should deliver (your user id, group id, or channel id; see below).

Optional:

- `LLM_*` — with **`LLM_API_KEY`** set and **`LLM_PROVIDER`** empty or **`openai`**, a normal **`make run`** uses the **OpenAI JSON digest** (`DigestSummarizer`, structured `picks` / `sections[].picks` → **Telegram HTML**). Without a key (or with a non-OpenAI provider), the run uses **`HeadlineSummarizer`** (same HTML layout, two links per subject when sectioned). The same variables power **`-llm-smoke`** (single ping, no RSS/Telegram).
- `NEUROFEED_LLM_MAX_ARTICLES` — max articles sent to the model after dedup (default **12**, allowed **1–40**). When feeds use **`subject`**, the batch is built **round-robin by subject** so early feeds (e.g. many AI sources) do not fill the whole cap and hide NBA/Tech/etc.; raise the cap if you want more depth per topic.
- `NEUROFEED_LLM_MAX_OUTPUT_TOKENS` — chat completion **max_tokens** for the digest (default **2500**, allowed **256–8192**).
- `NEUROFEED_RSS_ITEMS_PER_FEED` — per **RSS URL**, keep only the **N** newest entries by publication date (default **2**, allowed **0–50**; **0** = no per-feed cap).

When you run **`make run`** (without `-llm-smoke`), on success you should see structured logs on stderr including `step=pipeline_run_ok`. The bot receives one **`sendMessage`** payload with **`parse_mode: HTML`**: either an **LLM-filled digest** (when OpenAI is configured) or the **headline** path with the same section layout (up to **two** links per assunto, deduplicated upstream).

### LLM smoke (OpenAI HTTP, no RSS)

To verify `LLM_API_KEY` and network reachability without fetching feeds or Telegram:

```bash
export LLM_API_KEY="your-key"
# optional: LLM_BASE_URL LLM_MODEL LLM_PROVIDER NEUROFEED_LLM_TIMEOUT
go run ./cmd/neurofeed -llm-smoke
# or: make llm-smoke
```

On success, stderr shows `llm_smoke_ok` with a short model reply. RSS and Telegram env vars are **not** required for this path.

If configuration fails validation, the program logs an error and exits with code 1.

#### How to talk to your bot and get `TELEGRAM_CHAT_ID`

1. In Telegram, open **@BotFather** → `/newbot` (or use an existing bot) → copy the **HTTP API token** into `TELEGRAM_BOT_TOKEN`.
2. Open a chat with **your bot** in Telegram and send any message (for example `/start`). The bot must receive at least one message before `getUpdates` shows a `chat` object.
3. In a browser or with `curl`, call (replace `<token>`):

   `https://api.telegram.org/bot<token>/getUpdates`

4. In the JSON response, find `"chat":{"id": ... }` under `message` or `edited_message`. That numeric **id** is your `TELEGRAM_CHAT_ID` (groups often look like `-100...`).

After that, `make run` with env set will POST to Telegram’s `sendMessage` and you should see the digest in that same chat.

**How it works (short):** the binary fetches one or more feeds with `gofeed`, maps entries to `Article` (including `SourceTier` per feed), deduplicates by title, then either calls OpenAI for a **structured JSON digest** (assembled to **HTML** for Telegram) when `LLM_API_KEY` is configured for OpenAI, or builds the **same HTML layout** from titles/links without an LLM. Finally it calls the Telegram Bot API `sendMessage` (`parse_mode: HTML`) for your `chat_id`.

## Build a binary

```bash
make build
./bin/neurofeed
```

## Checks before you push

```bash
make fmt    # go fmt ./...
make vet    # go vet ./...
make test   # go test ./... -count=1
make all    # fmt, vet, test, build (in that order)
```

## Security reminder

Treat `TELEGRAM_BOT_TOKEN`, `LLM_API_KEY`, and feed URLs as secrets. Never commit real tokens or paste them into public issues or chat logs.
