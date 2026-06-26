# Implementation Plan — SI-TEPAT Backend

**Companion documents:** [`prd.md`](./prd.md) (product) · [`trd.md`](./trd.md) (technical + gap analysis)
**Purpose:** A prioritized, executable plan to evolve the working MVP into a complete,
production-ready system. Items are ordered **most critical (HIGH) → lowest (LOW)** and
map 1:1 to the `GAP-xx` IDs defined in the TRD.

**Effort key:** **S** ≈ ≤1 day · **M** ≈ 1–3 days · **L** ≈ 3–5 days (single engineer).
These are relative sizes, not commitments.

---

## 0. Master Backlog (priority order)

| # | Gap | Title | Priority | Effort | Depends on |
|---|---|---|---|---|---|
| 1 | GAP-04 | Quota period & timezone correctness | 🔴 HIGH | S | — |
| 2 | GAP-02 | User / officer management API | 🔴 HIGH | M–L | — |
| 3 | GAP-01 | Citizen read / search / list / detail | 🔴 HIGH | M | — |
| 4 | GAP-03 | Transaction monitoring: filters + pagination | 🔴 HIGH | M | — |
| 5 | GAP-07 | Production CORS + rotate seed creds + admin bootstrap | 🟠 MED-HIGH | S | — |
| 6 | GAP-06 | Rate limiting / brute-force protection on login | 🟠 MED-HIGH | S | — |
| 7 | GAP-05 | Automated tests (claim concurrency + RBAC) | 🟠 MED-HIGH | M | 1–4 |
| 8 | GAP-08 | Token refresh + revocation / `is_active` enforcement | 🟠 MED-HIGH | M | GAP-02 |
| 9 | GAP-10 | CI pipeline + observability | 🟡 MED | M | GAP-05 |
| 10 | GAP-09 | Cold-start latency mitigation | 🟡 MED | S | — |
| 11 | GAP-11 | OpenAPI / Swagger spec | 🟢 LOW | S | 1–4 |
| 12 | GAP-12 | Versioned migrations (golang-migrate) | 🟢 LOW | M | — |
| 13 | GAP-16 | Input hardening / central validation | 🟢 LOW | S | — |
| 14 | GAP-13 | Claim idempotency / dedupe window | 🟢 LOW | M | — |
| 15 | GAP-14 | Real-time monitoring (SSE/WebSocket) | 🟢 LOW | L | GAP-03 |
| 16 | GAP-15 | Confirm Vercel Go version support | 🟢 LOW | S | — |
| — | DOC-01 | Reconcile PRD §3.B wording with web-service model | 🟢 LOW | S | — |

> **Cross-cutting rule:** every feature item (1–4, 8) ships **with its tests** as part
> of Definition of Done. GAP-05 is the dedicated effort to backfill tests for the
> already-built claim path and to set up the test harness.

---

## Milestone 1 — Correctness & Functional Completeness 🔴 HIGH

**Goal:** the claim path is correct, and the admin dashboard can be fully built.
**Exit criteria:** dashboard can onboard officers, manage citizens & quotas, and
monitor transactions; no silent wrong-rejections from period mismatch.

### Task 1 · GAP-04 — Quota period & timezone correctness (S)
- **Why first:** this is a **correctness bug in the live claim flow** — legitimate
  claims can be silently rejected, and UTC month boundaries differ from WIB.
- **Scope:**
  - Add a single source of truth for the active period using **Asia/Jakarta** time,
    e.g. `models.CurrentPeriod() string` (format `YYYY-MM`).
  - Use it in `ClaimService.Claim` and as the default in `AdminHandler.SetQuota`.
  - Decide & document the strategy: claims resolve against the **current period**;
    admins set quotas for that period (and optionally future).
  - Files: `internal/services/claim_service.go`, `internal/handlers/admin_handler.go`,
    new helper in `internal/models`.
- **Acceptance:** a quota set for the active period is always honored; period is derived
  deterministically in WIB; behavior documented in TRD §7/§11.

### Task 2 · GAP-02 — User / officer management API (M–L)
- **Why:** officers currently exist only via `cmd/seed` — onboarding is impossible in
  production. Also unblocks GAP-08.
- **Scope:**
  - Schema: add `users.is_active bool not null default true` (migration via `cmd/migrate`).
  - Endpoints (admin-only):
    - `POST   /api/v1/admin/users` — create admin/merchant (bcrypt hash, `merchant_name`).
    - `GET    /api/v1/admin/users?role=&search=&page=&limit=` — list.
    - `PATCH  /api/v1/admin/users/:id` — update `merchant_name` / `role` / `is_active` / reset password.
  - Files: `internal/models/user.go`, `internal/repositories/user_repository.go`,
    new `internal/services/user_service.go` (or extend admin service),
    `internal/handlers/admin_handler.go`, routes in `internal/server`.
- **Acceptance:** an admin creates a new officer via API; that officer logs in;
  disabling (`is_active=false`) blocks future logins.

### Task 3 · GAP-01 — Citizen read / search / list / detail (M)
- **Why:** the dashboard cannot display or look up registered citizens today.
- **Scope:**
  - `GET /api/v1/admin/citizens?search=&page=&limit=` — paginated list + search by
    NIK / NFC UID / name.
  - `GET /api/v1/admin/citizens/:id` — detail including current-period quotas.
  - Files: `internal/repositories/citizen_repository.go` (+List/Search),
    `internal/repositories/quota_repository.go` (+ListByCitizen),
    `internal/services/admin_service.go`, `internal/handlers/admin_handler.go`, routes.
- **Acceptance:** dashboard searches a citizen by NIK and shows their quotas & status.

### Task 4 · GAP-03 — Transaction monitoring: filters + pagination (M)
- **Why:** PRD §3.C requires a real monitoring table; current endpoint only takes `limit`.
- **Scope:**
  - Extend `GET /api/v1/admin/transactions` with `page`, `from`, `to`, `status`,
    `commodity`, `user_id`, `merchant_name`; return `{ total, page, limit, data[] }`.
  - Files: `internal/repositories/transaction_repository.go` (query builder),
    `internal/services/admin_service.go`, `internal/handlers/admin_handler.go`.
- **Acceptance:** dashboard filters "today's rejected LPG claims for SPBU X" and paginates.

---

## Milestone 2 — Security & Reliability 🟠 MED-HIGH

**Goal:** safe to expose to real users and field officers.
**Exit criteria:** no default creds, brute-force protected, access revocable, core
logic covered by tests.

### Task 5 · GAP-07 — Production CORS + credential hygiene (S)
- **Scope:** require explicit `CORS_ALLOW_ORIGINS` in prod (no `*`); replace weak seed
  passwords with strong/generated values or a one-time bootstrap; document a
  "first admin" procedure. Files: deploy docs, `cmd/seed`, env guidance.
- **Acceptance:** prod restricts CORS to the dashboard origin; no default password works.

### Task 6 · GAP-06 — Login rate limiting / brute-force protection (S)
- **Scope:** add Fiber `limiter` (per-IP + per-username) on `/auth/login`; lockout/backoff
  after repeated failures; log attempts. Files: `internal/server`, maybe a small store.
- **Acceptance:** repeated failed logins are throttled and visible in logs.

### Task 7 · GAP-05 — Automated tests (M)
- **Scope:** set up the test harness; unit-test every `ClaimService` branch (not
  registered / not eligible / no quota / empty / success), validation tables, and RBAC;
  an **integration test firing N concurrent claims** proving `FOR UPDATE` (quota never
  negative). Files: `*_test.go` across services/handlers; a throwaway-Postgres harness.
- **Acceptance:** `go test ./...` green; concurrency test asserts no over-draw.

### Task 8 · GAP-08 — Token refresh + revocation (M, depends on GAP-02)
- **Scope:** short-lived access token + refresh token (or token-version/denylist);
  `RequireAuth` honors `is_active` per request. Files: `pkg/token`,
  `internal/middlewares`, `internal/services` (auth).
- **Acceptance:** disabling a user revokes access within the access-token TTL.

---

## Milestone 3 — Operations & Performance 🟡 MED

**Goal:** confident, observable deploys and predictable latency.

### Task 9 · GAP-10 — CI pipeline + observability (M, depends on GAP-05)
- **Scope:** GitHub Actions running `go vet` + `golangci-lint` + `go test` + `go build`
  on PRs; structured logging; error monitor (e.g. Sentry); basic metrics.
- **Acceptance:** PRs blocked on failing checks; runtime errors centralized.

### Task 10 · GAP-09 — Cold-start mitigation (S)
- **Scope:** production Neon plan without scale-to-zero and/or scheduled warm-up ping;
  document expected cold-start behavior and p95 targets.
- **Acceptance:** measured p95 (warm) meets target; cold-start path documented.

---

## Milestone 4 — Polish & Future 🟢 LOW

| Task | Gap | Scope summary |
|---|---|---|
| 11 | GAP-11 | OpenAPI/Swagger spec for mobile & web teams |
| 12 | GAP-12 | Replace `AutoMigrate` with versioned, reversible migrations |
| 13 | GAP-16 | Validate `nfc_uid` format; centralize validation (`go-playground/validator`) |
| 14 | GAP-13 | Optional idempotency key / short dedupe window on claims |
| 15 | GAP-14 | Live transaction feed via SSE/WebSocket (replaces polling) |
| 16 | GAP-15 | Confirm Vercel build image supports `go 1.26.4` (already deploys today; verify on upgrades) |
| — | DOC-01 | Update PRD §3.B wording: "Serverless Function" → "Vercel Go web service (Fluid Compute)" |

---

## Sequencing & Dependencies

```
Milestone 1 (HIGH):   GAP-04 → GAP-02 → GAP-01 → GAP-03   (mostly parallelizable; GAP-04 first as a correctness fix)
Milestone 2 (M-HIGH): GAP-07, GAP-06 (independent) ; GAP-05 (after features exist) ; GAP-08 (needs GAP-02)
Milestone 3 (MED):    GAP-10 (needs GAP-05) ; GAP-09 (independent)
Milestone 4 (LOW):    GAP-11..16, DOC-01 as capacity allows
```

**Quick wins (small, high value, do early):** GAP-04, GAP-07, GAP-06.

---

## Definition of Done (every task)

- Code follows existing layering (`handler → service → repository`) and Indonesian
  inline-comment style of the codebase.
- New/changed endpoints have request validation and consistent error envelopes
  (`{ "error": ... }` for technical errors).
- Tests added/updated; `go build ./...` and `go vet ./...` pass.
- TRD/PRD/this plan updated if a contract or scope changes.
- Deployed to Vercel and smoke-tested (`/health` + the new endpoint) before closing.

---

## Risks & Notes

- **Schema changes** (GAP-02 `is_active`) require running `cmd/migrate` against Neon
  before the new code is served — coordinate migration with deploy.
- **Concurrency tests** (GAP-05) need a real Postgres; SQLite cannot reproduce
  `SELECT … FOR UPDATE` semantics — use a disposable Neon branch or local Postgres.
- **Token revocation** (GAP-08) adds state; keep it lightweight (token version on the
  user row) to stay serverless-friendly.
- Keep the DB connection pool small (serverless) as new endpoints add query load.
