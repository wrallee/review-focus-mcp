# Review Focus MCP

Review Focus is an MCP App for **human pull-request review**. It reduces review noise by classifying changed files into attention buckets and rebuilding a focused Files Changed-style UI inside an MCP Apps-capable host.

- **SKIP**: low-value changes are folded by default.
- **REVIEW**: changes that deserve normal human review stay visible.
- **CRITICAL**: high-risk changes are expanded with why they matter, potential impact, and concrete review points.

The reviewer still owns the decision. Review Focus only changes where human attention is spent.

> MVP status: GitHub is implemented first. GitLab is intentionally kept behind the provider abstraction and is not implemented yet.

## Quick start with Docker

Docker is the recommended way to run Review Focus. You do not need Go, Node.js, npm, or `make` on the host.

### 1. Build the image

```bash
git clone https://github.com/wrallee/review-focus-mcp.git
cd review-focus-mcp
docker build -t review-focus-mcp .
```

### 2. Configure GitHub

GitHub.com and GitHub Enterprise Server use the **same environment variables**. Set the web URL for the GitHub instance you want to use:

```bash
export REVIEW_FOCUS_GITHUB_URL=https://github.com
export REVIEW_FOCUS_GITHUB_TOKEN=github_pat_xxx
```

For another GitHub instance, only the URL/token values change:

```bash
export REVIEW_FOCUS_GITHUB_URL=https://github.example.com
export REVIEW_FOCUS_GITHUB_TOKEN=YOUR_TOKEN
```

Review Focus derives the REST endpoint automatically:

- `https://github.com` → `https://api.github.com`
- any other GitHub web URL → `<REVIEW_FOCUS_GITHUB_URL>/api/v3`

Override it only when your environment exposes the API through a non-standard gateway/path:

```bash
export REVIEW_FOCUS_GITHUB_API_URL=https://gateway.example.com/github/api/v3
```

The REST API version header can also be overridden when needed:

```bash
export REVIEW_FOCUS_GITHUB_API_VERSION=2022-11-28
```

### 3. Run the MCP server

```bash
docker run --rm -i \
  -e REVIEW_FOCUS_GITHUB_URL \
  -e REVIEW_FOCUS_GITHUB_TOKEN \
  -e REVIEW_FOCUS_GITHUB_API_URL \
  -e REVIEW_FOCUS_GITHUB_API_VERSION \
  -e REVIEW_FOCUS_DATA_DIR=/data \
  -v review-focus-data:/data \
  review-focus-mcp
```

`-i` is required because the MCP server communicates over stdio. Do not add `-t`.

The named volume keeps local analysis and pending-review drafts across container restarts.

### Corporate/private CA

If the GitHub server uses a CA that is not in the container's default trust store, mount the CA and point Review Focus at it:

```bash
docker run --rm -i \
  -e REVIEW_FOCUS_GITHUB_URL \
  -e REVIEW_FOCUS_GITHUB_TOKEN \
  -e REVIEW_FOCUS_GITHUB_CA_FILE=/certs/company-ca.pem \
  -e REVIEW_FOCUS_DATA_DIR=/data \
  -v review-focus-data:/data \
  -v /absolute/path/company-ca.pem:/certs/company-ca.pem:ro \
  review-focus-mcp
```

Review Focus appends that certificate to the system trust pool. TLS verification is not disabled.

## Connect to Claude Code

After building the image, register the Docker command as a stdio MCP server:

```bash
claude mcp add review-focus -- \
  docker run --rm -i \
  -e REVIEW_FOCUS_GITHUB_URL \
  -e REVIEW_FOCUS_GITHUB_TOKEN \
  -e REVIEW_FOCUS_DATA_DIR=/data \
  -v review-focus-data:/data \
  review-focus-mcp
```

Verify the connection:

```bash
claude mcp list
claude mcp get review-focus
```

Inside Claude Code, `/mcp` also shows the connection status.

Example prompts:

```text
내가 지금 리뷰해야 하는 PR 보여줘
```

```text
Review Focus 열어줘
```

The `review_focus_open` tool is associated with the `ui://review-focus/main` MCP App resource. In an MCP Apps-capable host, the host renders the focused review UI from that resource.

## Human review flow

```text
Review requested PRs
       ↓
Select PR
       ↓
Analyze changes
       ↓
┌──────────┬──────────┬──────────┐
│   SKIP   │  REVIEW  │ CRITICAL │
│ folded   │ visible  │ expanded │
└──────────┴──────────┴──────────┘
       ↓
Human reads focused diff
       ↓
Add line comments
       ↓
Local pending draft
       ↓
COMMENT / REQUEST_CHANGES / APPROVE
       ↓
GitHub review submission
```

Before submit, Review Focus verifies that the PR HEAD SHA still matches the draft so comments are not silently submitted against stale code.

## MCP tools

| Tool | Purpose |
| --- | --- |
| `review_focus_open` | Open/load the current review-request inbox and MCP App |
| `review_focus_get` | Load PR metadata, changed files, cached analysis, and draft |
| `review_focus_analyze` | Classify files into SKIP / REVIEW / CRITICAL and cache by HEAD SHA |
| `review_focus_save_draft` | Persist the local pending human-review draft |
| `review_focus_submit` | Submit the GitHub review |
| `review_focus_health` | Check GitHub auth/connectivity and instance metadata |

## Development

Native build is intended for contributors. It requires Go 1.25+, Node.js 20+, and npm.

```bash
make build
make test
```

Equivalent commands without `make`:

```bash
cd ui
npm install
npm run build
npm run typecheck
cd ..
go test ./...
go build -o bin/review-focus-mcp ./cmd/review-focus
```

On Windows, build the Go executable as `bin/review-focus-mcp.exe`.

## Architecture

```text
Claude / MCP Apps host
        │ stdio
        ▼
Docker container / Go MCP server
  ├─ MCP tools + ui:// resource
  ├─ Analyzer interface
  ├─ local JSON store
  └─ SCM provider
       └─ GitHub adapter
              │ HTTPS
              ▼
        GitHub instance

React + TypeScript UI
        │ Vite single-file build
        ▼
internal/mcpapp/dist/index.html
        │ go:embed
        ▼
Go binary
```

The domain uses provider-neutral `ChangeRequest` types so a GitLab adapter can be added later without changing the review model.

See [`docs/design/ARCHITECTURE.md`](docs/design/ARCHITECTURE.md) and [`docs/design/SCM-COMPATIBILITY.md`](docs/design/SCM-COMPATIBILITY.md).

## Repository layout

```text
cmd/review-focus/          binary entry point
internal/analyzer/         change-attention classifier boundary
internal/config/           instance configuration
internal/domain/           provider-neutral review model
internal/mcpapp/           MCP tools, ui:// resource, embedded UI
internal/provider/github/  GitHub adapter
internal/storage/local/    MVP local JSON persistence
ui/                        React + TypeScript MCP App
docs/design/               product/architecture design
docs/adr/                  architecture decisions
```

## MVP boundaries

Not in the first version:

- GitLab implementation
- Production OAuth / multi-user auth
- Server-side database
- Distributed idempotency
- Multi-line comment selection
- Thread-resolution UI
- Full replacement for every GitHub PR tab

The first version proves the focused human-review loop before adding platform breadth.
