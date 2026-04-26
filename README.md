# Neurofeed

Neurofeed ingests RSS feeds, filters and ranks articles, optionally summarizes them with an LLM, and delivers a digest to Telegram. The service is written in Go with a small internal pipeline (`ingest` → `domain` → `ai` → `notify`).

## Development status

**Phases 0–1 are done** (repo scaffold, quality baseline, single RSS feed → Telegram digest). The runnable path matches **Phase 1** in the roadmap; a successful run logs `neurofeed phase 1 run OK`.

**Phases 2–7 are not done yet** (multi-source feeds, keyword scoring, richer LLM formatting, profiles, robustness). See the phase table in [`.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md`](.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md).

## Run locally

**[docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md)** — prerequisites, environment variables, `make run`, and Telegram setup.

Quick path: copy [`.env.example`](.env.example) to `.env`, fill in values, load env from the repo root, then `make run`.

## More documentation

- **[docs/README.md](docs/README.md)** — index of product spec, rules, and workflows
- **[AGENTS.md](AGENTS.md)** — instructions for AI coding agents
