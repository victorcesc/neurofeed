# Neurofeed

Neurofeed ingests RSS feeds, deduplicates articles, optionally summarizes them with an LLM, and delivers a digest to Telegram. The service is written in Go with a small internal pipeline (`ingest` → `domain` → `ai` → `notify`).

Phase checklist (what shipped vs next): **[SUMMARY.md](SUMMARY.md)**.

## Development status

**Phases 0–2 are done** (scaffold, single or multi RSS with per-feed **tiers**, title dedup, digest → Telegram). A successful run logs `step=pipeline_run_ok`.

**Phase 3 (3.1–3.3) is done** — OpenAI chat client, **`-llm-smoke`**, JSON digest summarizer wired on **`make run`** when `LLM_API_KEY` is set for OpenAI; caps via **`NEUROFEED_LLM_MAX_ARTICLES`** / **`NEUROFEED_LLM_MAX_OUTPUT_TOKENS`**. See [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md) and [SUMMARY.md](SUMMARY.md).

**Phases 4–6 are not done yet** (richer Telegram formatting, multi-recipient config, retries/logging/cache). See [`.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md`](.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md).

## Run locally

**[docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md)** — prerequisites, environment variables, `make run`, and Telegram setup.

Quick path: copy [`.env.example`](.env.example) to `.env`, fill in values, load env from the repo root, then `make run`.

## More documentation

- **[docs/README.md](docs/README.md)** — index of product spec, rules, and workflows
- **[AGENTS.md](AGENTS.md)** — instructions for AI coding agents
