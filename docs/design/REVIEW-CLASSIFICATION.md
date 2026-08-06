# Review Classification

## Principle

Classification answers **where a human should spend review attention**, not whether code is objectively safe.

Two axes are deliberately separated:

- `attention`: SKIP / REVIEW / CRITICAL
- `changeTypes`: semantic categories such as TRANSACTION or CACHE

## Attention levels

### SKIP

Default UI behavior: folded. Typical candidates are tests, documentation, generated or mechanical changes when confidently identified. `SKIP` must never be phrased as “safe” or “approved by AI.” A reviewer can always expand it.

### REVIEW

Default UI behavior: visible. Typical candidates are business logic, API behavior, DTO/data-shape changes, and ordinary implementation changes.

### CRITICAL

Default UI behavior: expanded and emphasized. The classifier should provide why the change is important, potential impact, and concrete review points.

Initial high-attention families include database schema/migrations, authorization/access control, transaction boundaries, concurrency/synchronization, and cache behavior.

## Analyzer contract

The analyzer is a replaceable boundary. `rules-v0` is intentionally deterministic and exists to validate the focused-review workflow. A later AI analyzer should preserve the same output contract and add evidence-grounded classification rather than generic AI review comments.
