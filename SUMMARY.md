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

**Done.** Multiple feeds via `NEUROFEED_RSS_FEEDS` (JSON) or legacy `RSS_FEED_URL` + optional `RSS_FEED_TIER`; each feed has a **source tier** (`primary` / `expert` / `news` / `community`) stamped on `Article.SourceTier`. `ingest.MultiRSSFetcher` aggregates feeds; **`NEUROFEED_RSS_ITEMS_PER_FEED`** (default **2**) keeps the **newest N items per feed URL** before merge; `domain.DeduplicateArticlesByTitle` after fetch; pipeline orchestrates fetch → dedup → summarize → notify.

---

## Phase 3 — AI digest (3.1–3.3)

**Done.**

- **3.1** — `internal/ai.OpenAIChatClient`; env `LLM_API_KEY`, `LLM_BASE_URL`, `LLM_MODEL`, `LLM_PROVIDER`, `NEUROFEED_LLM_TIMEOUT`; **`-llm-smoke`** / **`make llm-smoke`** for a minimal completion without RSS/Telegram.
- **3.2** — `DigestSummarizer`: system prompt + numbered article batch; **`response_format: json_object`**; parse, optional markdown-fence strip (`internal/ai/digest_*.go`). **Phase 4** narrowed the schema to structured **`picks`** (flat) or **`sections[].picks`** (per subject) before Telegram send.
- **3.3** — Env **`NEUROFEED_LLM_MAX_ARTICLES`** (default 12, max 40) and **`NEUROFEED_LLM_MAX_OUTPUT_TOKENS`** (default 2500, max 8192); **`make run`** uses **`DigestSummarizer`** when `LLM_API_KEY` is set and `LLM_PROVIDER` is empty or `openai`; otherwise **`HeadlineSummarizer`** (warn if provider is non-OpenAI). Extended run context deadline when the LLM path is active. **`httptest`** coverage for client JSON mode and digest summarizer.
- **RSS `subject`** — Optional **`subject`** per feed in **`NEUROFEED_RSS_FEEDS`** (or **`RSS_FEED_SUBJECT`** for single-feed): stamped on **`Article.Subject`**; when any feed has a subject, the LLM digest is **sectioned** by subject. Items from feeds without `subject` bucket as **`Geral`**.

---

## Phase 4 — Telegram message UX (HTML digests)

**Done.** Per **subject** (or implicit **`Geral`** when flat): up to **two** links per section, each with **two** summary lines (`line1` / `line2`) from the LLM JSON; **fallback** fills from the first articles with valid URLs if the model omits picks. **`internal/ai/digest_phase4.go`** assembles **Telegram HTML** (`<b>`, `🔗` + `<a href>`); **`notify.TelegramNotifier`** sends **`parse_mode: HTML`** (from `main`). **`HeadlineSummarizer`** uses the same layout without LLM lines (`notify.EscapeTelegramHTML` + escaped `href`). **`DigestSubjectSections`** (config feed order) plus **`EnrichSubjectOrderWithArticles`** and **`SubjectOrderedSummarizer`** ensure **every configured `subject`** appears each run (empty → placeholder). **`domain.BalanceArticlesAcrossSubjects`** round-robins the **`NEUROFEED_LLM_MAX_ARTICLES`** batch by subject so many AI feeds do not starve NBA/Tech/Futebol. Legacy `{"digest":"…"}` parser remains in **`parseDigestJSON`** for tests / compatibility only.

---

## Next phases (not implemented yet)

| Phase | Focus |
|-------|--------|
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
