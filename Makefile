.PHONY: ui build test fmt

ui:
	cd ui && npm install && npm run build

build: ui
	mkdir -p bin
	go build -o bin/review-focus-mcp ./cmd/review-focus

test:
	go test ./...
	cd ui && npm install && npm run typecheck

fmt:
	gofmt -w cmd internal
