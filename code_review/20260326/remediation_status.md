# Remediation Status — Synthetic Action Plan (2026-03-26)

Tracking document for the 12 findings in [`synthetic_action_plan.md`](./synthetic_action_plan.md).
The original plan is kept unmodified as a historical snapshot; all status updates live here.

> **Convention**: update this file whenever a finding is implemented or confirmed as a false positive.
> See workflow: `.agent/workflows/version-bump-and-changelog.md`

---

## Phase 1 — Critical Security & Data Integrity

| # | Finding | Priority | Status | Version | Resolution                                                                                                                                                                                                    |
|---|---------|----------|--------|---------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1 | Remove Credential Logging | Critical | ✅ Implemented | v0.3.2 | Removed debug log lines exposing password hashes in `login()`. Residual: test file still logs plaintext `adminPassword` at `Warn` — to be addressed.                                                          |
| 2 | Fix Soft-Delete Integrity | Critical | ✅ Implemented | v0.3.2 | Enforced soft-delete integrity in SQL queries by ensuring soft-deleted items are explicitly filtered and cannot be accessed/modified by ID.                                                                   |
| 3 | Restrict Over-Permissive RBAC | Critical | ⚠️ Partially Implemented | v0.3.2  | Replaced `system:serviceaccounts` group subject with a dedicated `go-cloud-k8s-thing-sa` ServiceAccount. `ClusterRoleBinding` still used — downgrade to namespace-scoped `RoleBinding` is a remaining action. |
| 4 | Enforce API Pagination Limits | High | ✅ Implemented | v0.3.2   | Enforced `MaxPaginationLimit = 1000` constant across all business layer  list/search queries to prevent unconstrained data fetching.         |

---

## Phase 2 — Operational Repairs (Quick-Wins)

| # | Finding | Priority | Status | Version | Resolution |
|---|---------|----------|--------|---------|------------|
| 5 | Fix Dockerfile HEALTHCHECK | High | ✅ Implemented | v0.3.4 | Removed `curl`-based `HEALTHCHECK` from `Dockerfile` (curl absent in scratch image). K8s probes handle health/readiness. |
| 6 | Decouple Liveness Probe from DB | High | ❌ False Positive / ⚠️ Partially Fixed | v0.3.4 | Liveness was already decoupled (returns `true` always). Real issue was `IsDBAlive()` using unbounded `context.Background()` — fixed to use `context.WithTimeout(2s)`. Comment added to `checkHealthy()` to prevent future regression. |
| 7 | Implement Graceful Shutdown | High | ❌ False Positive | — | Graceful shutdown fully implemented in `goHttpEcho.StartServer()` via `waitForShutdownToExit()` (handles `SIGTERM`/`SIGINT` with drain timeout). Missed by reviewers who only analysed `main()` without reading the library. Comment added to call site in `main()`. |

---

## Phase 3 — Reliability & Logic Correction

| # | Finding | Priority | Status | Version | Resolution |
|---|---------|----------|--------|---------|------------|
| 8 | Fix Proto3 Boolean Filtering | Medium | ✅ Implemented | v0.3.5 | Switched `bool` fields `validated` and `inactivated` to `optional bool` in search/list request definitions so that `false` is distinguishable from unset. |
| 9 | Stop Error Masking in Storage | Medium | ✅ Implemented | v0.3.5 | Refactored `Storage` methods (`Exist`, `IsThingActive`, `IsUserOwner`) to return `(bool, error)` to prevent swallowing underlying connection errors. |
| 10 | Correct Context Initialization | Medium | ❌ False Positive | — | The startup context is only used synchronously during init queries (`existTypeThing`). It is not retained by the module, so no background tasks are at risk. |

---

## Phase 4 — Refactoring & Technical Debt

| # | Finding | Priority | Status | Version | Resolution |
|---|---------|----------|--------|---------|------------|
| 11 | Optimize DB Operations | Low | 🔲 Pending | — | `Create`/`Update` do an extra `Get` round-trip. Use `RETURNING *` in SQL to eliminate it. |
| 12 | Clean Up Generated Code & Docs | Low | 🔲 Pending | — | `api/thing.yaml` is stale; `thing_types.gen.go` is no longer generated. Rename to `thing_types.go` and remove stale OpenAPI artefacts. |

---

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ Implemented | Fix applied and verified |
| ❌ False Positive | Finding incorrect — no change needed |
| ⚠️ Partially Implemented | Partially addressed; remaining action noted |
| 🔲 Pending | Not yet actioned |
