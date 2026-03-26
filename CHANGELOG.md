# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.2] - 2026-03-26

This release addresses the critical and high-priority findings (1 to 4) identified in Phase 1 of the [Synthetic Action Plan](code_review/20260326/synthetic_action_plan.md).

### Security
- **Credential Logging**: Removed debug statements that were logging password hashes during login (`cmd/goCloudK8sThingServer/goCloudK8sThingServer.go`).
- **RBAC**: Replaced the overly permissive cluster-wide `system:serviceaccounts` binding with a dedicated `go-cloud-k8s-thing-sa` ServiceAccount for the `pod-reader-role` and `service-reader-role` (`deployments/go-testing/deployment.yml`).

### Fixed
- **Soft-Delete Integrity**: Fixed SQL queries to append `AND _deleted = false` explicitly, ensuring that soft-deleted items cannot be accessed or manipulated directly by ID (`pkg/thing/thing_sql.go`).
- **API Pagination Limits**: Enforced a `MaxPaginationLimit` of 1000 across all `BusinessService` list and search endpoints to prevent unbounded queries from overloading the database or application memory (`pkg/thing/business_service.go`).
