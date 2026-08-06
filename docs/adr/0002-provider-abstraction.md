# ADR 0002: Provider-neutral change review domain

## Status
Accepted

## Decision

The core domain uses `ChangeRequest`, `ChangedFile`, `ReviewDraft`, and `ReviewComment`. GitHub-specific request/response models stay inside `internal/provider/github`.

GitLab will implement the same SCM interface later and translate GitLab diff-position semantics inside its adapter.
