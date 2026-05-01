# Neurofeed

Neurofeed ingests RSS feeds, deduplicates articles, optionally summarizes them with an LLM, and delivers a digest to Telegram. The service is written in Go with a small internal pipeline (`ingest` → `domain` → `ai` → `notify`).

Phase checklist (what shipped vs next): **[SUMMARY.md](SUMMARY.md)**.

## Development status

**Phases 0–2 are done** (scaffold, single or multi RSS with per-feed **tiers**, title dedup, digest → Telegram). A successful run logs `neurofeed run OK`.

**Phases 3–6 are not done yet** (LLM summaries, richer Telegram formatting, operator-defined multi-recipient config, retries/logging/cache). See the phase table in [`.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md`](.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md).

## Run locally

**[docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md)** — prerequisites, environment variables, `make run`, and Telegram setup.

Quick path: copy [`.env.example`](.env.example) to `.env`, fill in values, load env from the repo root, then `make run`.

## More documentation

- **[docs/README.md](docs/README.md)** — index of product spec, rules, and workflows
- **[AGENTS.md](AGENTS.md)** — instructions for AI coding agents
