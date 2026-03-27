---
description: how to bump the version and update the changelog after implementing fixes
---

## Context

- Version is defined in `pkg/version/version.go` as the `VERSION` variable.
- Changelog is `CHANGELOG.md` at the repository root, following [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format and [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
- Code review findings are tracked in `code_review/20260326/remediation_status.md`.

## Versioning Rules

Use **patch version** bumps (x.y.**Z**) for:
- Bug fixes, security patches, operational improvements, documentation-only changes.

Use **minor version** bumps (x.**Y**.0) for:
- New features, new endpoints, new module capabilities, non-breaking API additions.

Use **major version** bumps (**X**.0.0) for:
- Breaking API changes, major architectural refactors.

## Steps

1. Identify the current version by reading `pkg/version/version.go`.

// turbo
2. Increment the appropriate version segment in `pkg/version/version.go`.

// turbo
3. Prepend a new section to `CHANGELOG.md` using this template:

```markdown
## [X.Y.Z] - YYYY-MM-DD

This release addresses <brief description of scope>.

### Security        ← include only if relevant
- **<Topic>**: <What was done and why>. (`<file path>`)

### Fixed           ← include only if relevant
- **<Topic>**: <What was done and why>. (`<file path>`)

### Added           ← include only if relevant
- **<Topic>**: <What was done and why>. (`<file path>`)

### Documentation   ← include only if relevant
- **<Topic>**: <What was done and why>. (`<file path>`)
```

// turbo
4. If the changes resolve one or more findings from a code review action plan, update `code_review/20260326/remediation_status.md`:
   - Set the finding's **Status** to `✅ Implemented in vX.Y.Z` or `❌ False Positive` as appropriate.
   - Fill in the **Resolution** column with a one-line description.

// turbo
5. Run `go build ./...` from the repository root to confirm the build is clean before committing.

6. Propose a commit message following the project convention (see example below).
   Present it as a ready-to-copy block so the user can commit directly.

   **Format rules:**
   - **Subject line**: `<type>: <imperative short description> (<scope>)` — max 72 chars.
     - `type` is one of: `fix`, `feat`, `refactor`, `chore`, `docs`, `test`, `ci`.
     - Use `fix:` for bug-fix/security/operational patches; `feat:` for new capabilities.
   - **Blank line** between subject and body.
   - **Body**: 3–6 bullet lines starting with `-`, each describing one concrete change in past tense.
     - Reference the specific file or symbol changed where helpful.
   - **Blank line** then a `Refs:` trailer pointing to the relevant action plan or issue.

   **Template:**
   ```
   <type>: <short imperative description> (<scope>)

   This commit addresses <findings / feature description>.

   Changes:
   - <Concrete change 1, past tense, with file/symbol reference if useful.>
   - <Concrete change 2.>
   - <Concrete change 3.>

   Refs: <e.g. code_review/20260326/synthetic_action_plan.md>
   ```

   **Example** (from commit `e34e88ce`, tag `v0.3.2`):
   ```
   fix: mitigate critical and high-priority sec/ops findings (Phase 1)

   This commit addresses Findings 1-4 identified in the recent Phase 1
   code review synthetic action plan to improve security, data integrity,
   and cluster stability.

   Changes:
   - Removed credential/password hash logging from the authentication flow.
   - Enforced soft-delete integrity in SQL queries by ensuring soft-deleted
     items are explicitly filtered and cannot be accessed/modified by ID.
   - Restricted over-permissive cluster-wide RBAC role bindings by creating
     a dedicated namespace-bound ServiceAccount (`go-cloud-k8s-thing-sa`).
   - Enforced a `MaxPaginationLimit` of 1000 across all business layer
     list/search queries to prevent unconstrained data fetching.
   - Initialized standard CHANGELOG.md and bumped application version to 0.3.2.

   Refs: code_review/20260326/synthetic_action_plan.md
   ```
