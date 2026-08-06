# Architecture

## Decision

Use a Go MCP server and a React/TypeScript MCP App bundled as a single HTML resource and embedded into the Go binary.

## Boundaries

```text
MCP transport
    ↓
MCP application adapter
    ↓
Domain + review use cases
 ┌───────┼──────────┐
 ↓       ↓          ↓
SCM    Analyzer    Store
 ↓                   ↓
GitHub              Local JSON
GitLab(future)      Remote store(future)
```

The domain package must not import GitHub, MCP, React, filesystem, or HTTP-specific types.

## Local MVP

Claude starts `review-focus-mcp` as a child process. The MCP server communicates over stdio. The UI is returned as `ui://review-focus/main` with `text/html;profile=mcp-app` and rendered by an MCP Apps-capable host.

Local state lives under `~/.review-focus` and is keyed by SCM identity + PR number + HEAD SHA.

## Remote migration

Remote deployment must not require changing the domain or provider contracts. Replace only:

- stdio transport → Streamable HTTP transport
- local token/config → authenticated user/tenant context
- local JSON store → durable server-side store
- local idempotency → distributed idempotency

The remote phase must introduce OAuth/user identity before any write operation is shared by multiple users.

## Provider abstraction

GitHub `PullRequest` and GitLab `MergeRequest` are represented as provider-neutral `ChangeRequest` values. Provider-specific review anchors are converted inside each adapter.
