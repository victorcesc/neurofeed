# Neurofeed — delivery summary

High-level status of roadmap phases ([product spec](docs/neurofeed.md), [engineering plan](.cursor/plans/neurofeed_go_plan_2338e2b3.plan.md)). For how to run and env vars, see [docs/HOW_TO_RUN.md](docs/HOW_TO_RUN.md).

---

## Phase 0 — Repo and quality baseline

**Done.** Runnable Go module, layout under `cmd/` and `internal/`, `context` on I/O boundaries, handbooks (`docs/RULES.md`, `docs/SKILLS.md`, `docs/neurofeed.md`, `AGENTS.md`), `Makefile` (`fmt`, `vet`, `test`, `run`, `build`). Stub fetchers / notifier / summarizer wired with interfaces for tests.

---

## Phase 1 — MVP (one path end-to-end)

**Done.** Single RSS URL from env (`RSS_FEED_URL`), map items to `domain.Article`, build a plain-text digest (title + link) via `ai.HeadlineSummarizer`, send with `notify.TelegramNotifier` (`TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`). Config validation for that path.

---

## Phase 2 — Multiple sources

**Done.** Multiple feeds via `NEUROFEED_RSS_FEEDS` (JSON) or legacy `RSS_FEED_URL` + optional `RSS_FEED_TIER`; each feed has a **source tier** (`primary` / `expert` / `news` / `community`) stamped on `Article.SourceTier`. `ingest.MultiRSSFetcher` aggregates feeds; `domain.DeduplicateArticlesByTitle` after fetch; pipeline orchestrates fetch → dedup → summarize → notify.

---

## Phase 3.1 — OpenAI HTTP client (smoke)

**Done.** `internal/ai.OpenAIChatClient` — OpenAI-compatible `POST …/chat/completions` with `context`, `Authorization: Bearer`, `User-Agent`. Config: `LLM_API_KEY`, `LLM_BASE_URL` (default `https://api.openai.com/v1`), `LLM_MODEL` (default `gpt-4o-mini`), `LLM_PROVIDER` (`openai` or empty for smoke), `NEUROFEED_LLM_TIMEOUT` (default `60s`) for the dedicated LLM HTTP client. CLI: **`go run ./cmd/neurofeed -llm-smoke`** or **`make llm-smoke`** validates I/O without running the RSS pipeline. Normal **`make run`** still uses `HeadlineSummarizer` until phase 3.3.

---

## Next phases (not implemented yet)

| Phase | Focus |
|-------|--------|
| **3.2** | Digest prompts from spec, article batch → request, structured / parseable model output |
| **3.3** | Token/article caps, output validation, wire real `Summarizer` in pipeline + `httptest` coverage |
| **4** | Telegram message UX: categories, emojis, Markdown, safe links |
| **5** | Operator-defined recipients/subjects/feeds; Telegram receive-only; per-subject **fixed curated RSS lists** (no LLM feed discovery); tier overrides |
| **6** | Retries/backoff, structured logging polish, caches, hardening |

---

## Rule — updating this file when a phase ships

When **Phase N** is completed in code (merged to the branch you treat as source of truth):

1. **Edit `SUMMARY.md`** in the same change (or immediately after), not only the plan file.
2. Move the phase from **Next phases** into a **Done** section with a short bullet list of what shipped (packages, env vars, user-visible behavior).
3. Keep the **Next phases** table accurate: remove or shrink the row for N; adjust wording if scope shifted.
4. If behavior or env vars changed, cross-check **`.env.example`**, **`docs/SKILLS.md`**, and **`docs/HOW_TO_RUN.md`** and update them in the same delivery when relevant.

This keeps `SUMMARY.md` a quick stakeholder and onboarding view; the detailed checklist stays in [docs/neurofeed.md](docs/neurofeed.md) and the plan under `.cursor/plans/`.

---

## Runtime logging (step-by-step)

On `make run` / `go run ./cmd/neurofeed`, **structured logs** go to **stderr** (`log/slog`, text handler). Useful attributes:

| Key | Meaning |
|-----|--------|
| `step` | Logical step name (e.g. `config_loaded`, `feed_fetch_start`, `dedup_done`, `notify_done`) |
| `detail` | Short human hint where it helps |

**Prefix-style grep:** lines include `neurofeed` (CLI), `pipeline` (orchestration), or `ingest` (per-feed RSS). Example: `make run 2>&1 | rg 'step='`.
