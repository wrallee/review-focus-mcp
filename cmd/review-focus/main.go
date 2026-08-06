package main

import (
	"context"
	"log"

	"github.com/wrallee/review-focus-mcp/internal/analyzer"
	"github.com/wrallee/review-focus-mcp/internal/config"
	"github.com/wrallee/review-focus-mcp/internal/mcpapp"
	githubprovider "github.com/wrallee/review-focus-mcp/internal/provider/github"
	"github.com/wrallee/review-focus-mcp/internal/storage/local"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	httpClient, err := cfg.HTTPClient()
	if err != nil {
		log.Fatal(err)
	}
	if err := mcpapp.EnsureEmbeddedUI(); err != nil {
		log.Fatal(err)
	}
	scm := githubprovider.New(cfg, httpClient)
	server := mcpapp.New(mcpapp.Service{SCM: scm, Analyzer: analyzer.Rules{}, Store: local.New(cfg.DataDir)})
	if err := mcpapp.RunStdio(context.Background(), server); err != nil {
		log.Fatal(err)
	}
}
