# ADR 0001: Go server with embedded React MCP App

## Status
Accepted

## Context

The local MVP must be easy to install as one executable, while the UI must use the MCP Apps browser environment. The server will later move to a remote runtime.

## Decision

- Go for MCP server, SCM clients, storage, and transport.
- React + TypeScript for the MCP App.
- Vite + `vite-plugin-singlefile` produces one HTML artifact.
- `go:embed` packages that HTML into the Go executable.
- stdio is the MVP transport; the application layer is transport-independent.

## Consequences

The project has two languages, but their boundary is explicit MCP/JSON contracts. Deployment remains one Go executable for the local MVP.
