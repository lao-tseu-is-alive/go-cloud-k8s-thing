# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.4] - 2026-03-27

This release addresses operational findings 5 and 6 from Phase 2 of the [Synthetic Action Plan](code_review/20260326/synthetic_action_plan.md).

### Fixed
- **Dockerfile HEALTHCHECK (Finding 5)**: Removed the broken `HEALTHCHECK` directive that used `curl` inside a `FROM scratch` image where `curl` does not exist. The check always failed silently. Health and readiness are managed exclusively by Kubernetes probes (`/health` and `/readiness`). (`Dockerfile`)
- **Readiness Probe Timeout (Finding 6)**: `IsDBAlive()` used `context.Background()` (unbounded), which could cause readiness probe goroutines to pile up if the database hung. Replaced with a 2-second timeout context. (`cmd/goCloudK8sThingServer/goCloudK8sThingServer.go`)

### Documentation
- **Liveness Probe Intent**: Added an explanatory comment to `checkHealthy()` documenting why the DB check is intentionally absent from the liveness probe — coupling liveness to DB state causes thundering-herd pod restart cascades on transient DB blips. Readiness (`checkReady`) remains the sole probe that checks DB connectivity. (`cmd/goCloudK8sThingServer/goCloudK8sThingServer.go`)

## [0.3.3] - 2026-03-26


### Fixed
- **Test Stability**: Fixed an invalid SQL syntax issue (`500 Internal Server Error`) in `countThing` queries caused by duplicate `WHERE` clauses, restoring full test suite stability (`pkg/thing/storage_postgres.go`, `pkg/thing/thing_sql.go`).

## [0.3.2] - 2026-03-26

This release addresses the critical and high-priority findings (1 to 4) identified in Phase 1 of the [Synthetic Action Plan](code_review/20260326/synthetic_action_plan.md).

### Security
- **Credential Logging**: Removed debug statements that were logging password hashes during login (`cmd/goCloudK8sThingServer/goCloudK8sThingServer.go`).
- **RBAC**: Replaced the overly permissive cluster-wide `system:serviceaccounts` binding with a dedicated `go-cloud-k8s-thing-sa` ServiceAccount for the `pod-reader-role` and `service-reader-role` (`deployments/go-testing/deployment.yml`).

### Fixed
- **Soft-Delete Integrity**: Fixed SQL queries to append `AND _deleted = false` explicitly, ensuring that soft-deleted items cannot be accessed or manipulated directly by ID (`pkg/thing/thing_sql.go`).
- **API Pagination Limits**: Enforced a `MaxPaginationLimit` of 1000 across all `BusinessService` list and search endpoints to prevent unbounded queries from overloading the database or application memory (`pkg/thing/business_service.go`).
