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

## Next phases (not implemented yet)

| Phase | Focus |
|-------|--------|
| **3** | Keyword positive/negative scoring, tier weights, sort, top N |
| **4** | OpenAI (or other) summarization with timeouts and prompts |
| **5** | Telegram message UX: categories, emojis, Markdown, safe links |
| **6** | Profiles, interest topics, tier overrides, feed subsets |
| **7** | Retries/backoff, structured logging polish, caches, hardening |

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
