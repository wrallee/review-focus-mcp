package config

import "testing"

func TestDerivesGHESAPIURL(t *testing.T) {
	t.Setenv("REVIEW_FOCUS_GITHUB_URL", "https://ghe.example.com/")
	t.Setenv("REVIEW_FOCUS_GITHUB_TOKEN", "x")
	t.Setenv("REVIEW_FOCUS_GITHUB_API_URL", "")
	t.Setenv("REVIEW_FOCUS_DATA_DIR", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GitHubAPIURL != "https://ghe.example.com/api/v3" {
		t.Fatalf("got %s", cfg.GitHubAPIURL)
	}
}
