# MVP Product Scope

## Goal

Reduce the amount of code a human reviewer must actively inspect without removing human ownership of the review decision.

## Core loop

1. Show pull requests that request the current user's review.
2. Load the selected PR and its changed files.
3. Classify each file into `SKIP`, `REVIEW`, or `CRITICAL`.
4. Fold `SKIP` changes by default.
5. Explain why `CRITICAL` changes deserve attention and what to verify.
6. Let the reviewer write inline comments while reading the focused diff.
7. Save those comments as a local pending draft.
8. Submit one GitHub review as `COMMENT`, `REQUEST_CHANGES`, or `APPROVE`.
9. Refuse submission if the PR HEAD SHA changed.

## Non-goals

- Replacing the reviewer with AI.
- Generating implementation/test specifications.
- Re-implementing every GitHub PR feature.
- GitLab support in the first vertical slice.
- Production multi-user state or distributed idempotency.

## Attention semantics

`SKIP` does not mean "proven safe". It means the change does not require focused human inspection under the active classification policy.

`REVIEW` means normal human inspection.

`CRITICAL` means the UI must show the change expanded together with the classification reason, possible impact, and concrete review points.
