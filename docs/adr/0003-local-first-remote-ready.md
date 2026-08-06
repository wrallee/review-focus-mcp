# ADR 0003: Local-first, remote-ready transport and storage

## Status
Accepted

## Decision

Start with a local stdio MCP server and local JSON persistence. Keep application logic independent of transport and storage so the server can later run remotely over Streamable HTTP.

## Rationale

The MVP optimizes iteration speed. Production remote deployment introduces identity, durable persistence, idempotency, tenancy, and operational concerns that should be added only after the review UX is validated.

## Consequence

HTTP mode is development-only until authentication and durable state are implemented.
