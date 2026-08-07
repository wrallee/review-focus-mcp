FROM node:22-bookworm AS ui
WORKDIR /src

COPY ui/package.json ./ui/package.json
RUN cd ui && npm install

COPY ui ./ui
RUN cd ui && npm run build

FROM golang:1.25-bookworm AS server
WORKDIR /src

COPY go.mod go.sum ./
COPY vendor ./vendor

COPY cmd ./cmd
COPY internal ./internal
COPY --from=ui /src/internal/mcpapp/dist/index.html ./internal/mcpapp/dist/index.html

RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/review-focus-mcp ./cmd/review-focus

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=server /out/review-focus-mcp /usr/local/bin/review-focus-mcp

ENV REVIEW_FOCUS_DATA_DIR=/data
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/review-focus-mcp"]
