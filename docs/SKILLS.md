# Neurofeed — workflows and procedures

[Index of all docs](README.md). Short guide for humans and coding agents. Coding standards live in [RULES.md](RULES.md).

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
```

Optional:

- `NEUROFEED_HTTP_TIMEOUT` — Go duration string (e.g. `45s`) for default HTTP client timeout baseline.
- `NEUROFEED_HTTP_TIMEOUT_SECONDS` — integer seconds; if set after the duration env is parsed, it overrides the timeout (see `internal/config`).

## Adding an RSS source (after phase 2)

1. Add the feed URL to configuration (env or future config file).
2. Confirm fetch timeout and `User-Agent` in the ingest package.
3. Run `make test` and a manual `make run` in a safe environment.

## Scheduler options (later)

- **OS cron / systemd timer**: invoke `bin/neurofeed` on a schedule; simplest ops model.
- **In-process gocron**: embed scheduling in the binary; fewer moving parts on a single host.

Document the chosen approach here once the project picks one.

## Prompt templates (phase 4)

Prompts from the product brief live alongside the LLM client abstraction under `internal/ai/` (to be added). To A/B prompts, use configuration to select template name or file path.

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
