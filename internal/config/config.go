package config

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	GitHubURL        string
	GitHubAPIURL     string
	GitHubAPIVersion string
	GitHubToken      string
	GitHubCAFile     string
	DataDir          string
}

func Load() (Config, error) {
	web := strings.TrimRight(env("REVIEW_FOCUS_GITHUB_URL", "https://github.com"), "/")
	api := strings.TrimRight(os.Getenv("REVIEW_FOCUS_GITHUB_API_URL"), "/")
	if api == "" {
		if web == "https://github.com" {
			api = "https://api.github.com"
		} else {
			api = web + "/api/v3"
		}
	}
	dataDir := os.Getenv("REVIEW_FOCUS_DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("resolve home directory: %w", err)
		}
		dataDir = filepath.Join(home, ".review-focus")
	}
	cfg := Config{
		GitHubURL:        web,
		GitHubAPIURL:     api,
		GitHubAPIVersion: env("REVIEW_FOCUS_GITHUB_API_VERSION", "2022-11-28"),
		GitHubToken:      os.Getenv("REVIEW_FOCUS_GITHUB_TOKEN"),
		GitHubCAFile:     os.Getenv("REVIEW_FOCUS_GITHUB_CA_FILE"),
		DataDir:          dataDir,
	}
	if _, err := url.ParseRequestURI(cfg.GitHubAPIURL); err != nil {
		return Config{}, fmt.Errorf("invalid REVIEW_FOCUS_GITHUB_API_URL: %w", err)
	}
	if cfg.GitHubToken == "" {
		return Config{}, errors.New("REVIEW_FOCUS_GITHUB_TOKEN is required")
	}
	return cfg, nil
}

func (c Config) HTTPClient() (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if c.GitHubCAFile != "" {
		pemBytes, err := os.ReadFile(c.GitHubCAFile)
		if err != nil {
			return nil, fmt.Errorf("read corporate CA: %w", err)
		}
		pool, err := x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		if ok := pool.AppendCertsFromPEM(pemBytes); !ok {
			return nil, errors.New("corporate CA file contains no certificates")
		}
		transport.TLSClientConfig = transport.TLSClientConfig.Clone()
		transport.TLSClientConfig.RootCAs = pool
	}
	return &http.Client{Transport: transport, Timeout: 30 * time.Second}, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
