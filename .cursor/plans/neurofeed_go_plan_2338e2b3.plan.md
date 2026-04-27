---
name: Neurofeed Go Plan
overview: Turn [docs/neurofeed.md](../../docs/neurofeed.md) into a phased delivery plan for a Go-based RSS→dedup→LLM→Telegram pipeline, with handbooks under docs/ (RULES, SKILLS) plus AGENTS.md at repo root for agent entry.
todos:
  - id: docs-rules-skills-agents
    content: Author docs/RULES.md, docs/SKILLS.md, docs/HOW_TO_RUN.md, docs/neurofeed.md; AGENTS.md at repo root (Go idioms, workflows, agent boundaries)
    status: completed
  - id: phase-0-scaffold
    content: Add go.mod, cmd/neurofeed, internal/* skeleton, config from env, context+timeouts on I/O stubs
    status: completed
  - id: phase-1-mvp
    content: Telegram send + single RSS feed + Article mapping + one-shot run
    status: completed
  - id: phase-2-6
    content: Iterate multi-source, OpenAI, formatting, profiles, retries/logging/cache per phase table
    status: pending
isProject: false
---

# Neurofeed: development phases and agent documentation

## Context

The product vision and technical stack are defined in [docs/neurofeed.md](../../docs/neurofeed.md): Go, RSS via [gofeed](https://github.com/mmcdole/gofeed), scheduling (cron or [gocron](https://github.com/go-co-op/gocron)), `net/http` for APIs, OpenAI for summaries, Telegram Bot API for delivery. Success is measured by signal quality (dedup + summarize), not raw volume.

## Target architecture (packages)

Keep boundaries explicit so each phase has a clear home and tests stay small:

```mermaid
flowchart LR
  subgraph ingest [ingest]
    RSS[RSS_fetch]
  end
  subgraph domain [domain]
    Article[Article_model]
    Dedup[dedup]
  end
  subgraph ai [ai]
    OpenAI[OpenAI_client]
  end
  subgraph out [out]
    TG[Telegram_sender]
    Format[message_format]
  end
  cron[cron_or_gocron] --> RSS
  RSS --> Article
  Article --> Dedup
  Dedup --> OpenAI
  OpenAI --> Format
  Format --> TG
```

Suggested layout (to implement after plan approval, not in this planning step):

- `cmd/neurofeed/` — thin `main`: parse flags/env, wire deps, run job once or on schedule
- `internal/config/` — validated configuration (feeds, tokens, timeouts, model name)
- `internal/ingest/` — RSS → `Article` (timeouts, user-agent, errors)
- `internal/domain/` — `Article`, dedup key normalization, source tiers
- `internal/ai/` — provider-agnostic LLM client(s), prompt templates from [docs/neurofeed.md](../../docs/neurofeed.md) (base + advanced)
- `internal/notify/` — Telegram send, Markdown/HTML escaping per [Telegram Bot API](https://core.telegram.org/bots/api) rules
- `internal/app/` or `internal/pipeline/` — orchestration: collect → dedup → summarize → format → send

Use **interfaces at boundaries** (e.g. `FeedFetcher`, `Summarizer`, `Notifier`) with concrete implementations in the same or adjacent packages so tests can use fakes without global state.

---

## Development phases

Phases below **merge** the roadmap in [docs/neurofeed.md](../../docs/neurofeed.md) with engineering work (config, structure, quality) so “maximum Go” applies from day one.

| Phase | Goal | Outcomes |
|-------|------|----------|
| **0 — Repo and quality baseline** | Runnable module, conventions locked | `go.mod`, minimal `main`, `.gitignore`, `docs/RULES.md` / `docs/SKILLS.md` / `AGENTS.md` (this plan’s deliverables), optional `Makefile` or `task` for `fmt`, `vet`, `test`. No feature code without `context.Context` on I/O boundaries. |
| **1 — MVP** | One path end-to-end | Telegram bot created; send one test message; fetch **one** RSS URL with gofeed; map to `Article`; send title + link to Telegram. Config via env (bot token, feed URL). |
| **2 — Multiple sources** | Scale ingestion | Multiple feeds from config (YAML/JSON or env list) **each with a source tier** (`primary` / `expert` / `news` / `community` per [docs/neurofeed.md](../../docs/neurofeed.md)); map to `Article` including `SourceTier`; dedup by normalized title (lower, strip punctuation; optional hash later). |
| **3 — AI integration** | Summaries | Split into **3.1–3.3** below (OpenAI HTTP + config; prompts and output shape; validation, caps, pipeline wiring). |
| **4 — Message UX** | Readable digest | Categories, emojis, Markdown (or HTML) with Telegram-safe formatting; clickable links. |
| **5 — Personalization** | Multi-audience | Profiles; **up to 5 interest topics** per profile (Telegram UX: **catalog + search** as primary, optional limited free-text per [docs/neurofeed.md](../../docs/neurofeed.md)); map interests → keyword/synonym lists; **tier weight overrides**; per-profile feed subsets. |
| **6 — Robustness** | Production habits | Retries with backoff for transient HTTP failures, structured logging (`log/slog`), request timeouts everywhere, simple TTL cache if needed to avoid duplicate API work. |

### Phase 3 sub-steps (3.1–3.3)

| Sub-step | Goal | Outcomes |
|----------|------|----------|
| **3.1** | HTTP + config | Env-backed OpenAI (or compatible) settings: API key, model, timeouts, optional base URL; `internal/ai` HTTP client using `context.Context` on every request; wire a minimal completion path (e.g. one fixed test prompt) to prove I/O without full digest logic. |
| **3.2** | Prompts + output shape | Prompt text aligned with [docs/neurofeed.md](../../docs/neurofeed.md) for the daily digest; map `[]domain.Article` → request body; structured output (JSON mode or strict parse) so the pipeline gets a stable string (or intermediate struct) for formatting/notify. |
| **3.3** | Validation + integration | Caps on articles/tokens per run (cost control from plan risks); simple heuristics or schema checks on model output; `Summarizer` implementation selected in `main`/config (e.g. OpenAI vs `HeadlineSummarizer` fallback); table-driven tests + `httptest` for the client. |

**Go-specific emphasis across phases**

- **Errors**: `%w` wrapping, sentinel errors where appropriate, no silent `_` on I/O.
- **Context**: every outbound HTTP and LLM call accepts `context.Context`.
- **Types**: small structs, constructor functions for clients, avoid `interface{}` unless justified.
- **Testing**: table-driven tests for dedup/tier parsing/formatting; `httptest` for HTTP clients; inject clocks when behavior depends on time.
- **Concurrency**: if parallel feed fetch, use bounded worker pattern or `errgroup` with context cancellation; document limits.
- **Documentation**: package comments on `internal/*` roots; exported symbols documented only where API is non-obvious.

---

## docs/RULES.md (planned contents)

Single source of truth for **how we write Go in this repo** (file: [docs/RULES.md](../../docs/RULES.md)):

- Module layout (`cmd/` vs `internal/`), naming, and when to add a new package.
- Error handling, logging (`slog`), configuration (env + optional file), secrets (never committed).
- HTTP: timeouts, default transport considerations, retries policy (where and max attempts).
- Telegram and LLM providers: rate limits, message length split strategy if digest exceeds limits.
- Testing expectations (what must be tested per change type), `go vet` / `staticcheck` if adopted.
- Dependencies: prefer stdlib + minimal deps; justify new modules.

Optional later: mirror critical bullets into [`.cursor/rules/*.mdc`](https://cursor.com/docs) for editor-native hints; **docs/RULES.md** remains the canonical human-readable contract.

---

## docs/SKILLS.md (planned contents)

**Index of repeatable workflows** for humans and agents (file: [docs/SKILLS.md](../../docs/SKILLS.md)):

- How to add a new RSS source and redeploy/run locally.
- How to tune prompts and digest limits via configuration when those knobs exist.
- How to run the daily job manually vs on schedule.
- How to rotate Telegram/OpenAI tokens safely.
- Pointers to prompt templates location and how to A/B base vs advanced prompt.
- “Definition of done” checklist per phase (maps to success criteria in [docs/neurofeed.md](../../docs/neurofeed.md)).

If you later add real [Cursor Agent Skills](https://cursor.com/docs) under `.cursor/skills/`, **docs/SKILLS.md** should link to those paths and one-line descriptions.

---

## AGENTS.md (planned contents)

**Instructions for AI coding agents** working in this repository:

- Read **docs/RULES.md** first; follow **docs/SKILLS.md** for operational steps (see [AGENTS.md](../../AGENTS.md) at repo root).
- Scope: prefer minimal diffs; match existing patterns; do not add unsolicited docs beyond what the team requests.
- Language: Go version target (state explicit version in AGENTS.md once chosen, e.g. 1.22+).
- Testing: run/format commands the project standardizes on.
- Security: never echo secrets; use env placeholders in examples.
- When to propose new packages vs extending an existing one (tie to architecture diagram above).

This complements Cursor’s native rules: **AGENTS.md** is repo-level “agent README”; **docs/RULES.md** is coding law; **docs/SKILLS.md** is procedure.

---

## Implementation order (after you approve)

1. Add **docs/RULES.md**, **docs/SKILLS.md**, **AGENTS.md** (repo root) with the sections above, tuned to your chosen Go version and scheduler (cron binary vs gocron in-process).
2. Scaffold **Phase 0** (`go.mod`, `cmd/neurofeed`, `internal/config`, minimal pipeline stub).
3. Execute phases **1–6** sequentially, merging “robustness” practices (timeouts, context) from phase 0 onward rather than deferring all to phase 6.

---

## Risks and decisions (non-blocking for the plan)

- **Telegram message length**: plan chunking or “continued” messages early in phase 4.
- **OpenAI cost**: cap articles per run and token limits in phase 3 config.
- **Scheduler**: systemd/cron wrapping a binary vs embedded gocron affects deployment docs in **docs/SKILLS.md**—call out both options there until you pick one.

After plan approval, the first concrete edits were the handbooks under **docs/** plus **AGENTS.md** and Phase 0 scaffold.
