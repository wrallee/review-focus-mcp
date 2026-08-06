package github

import "testing"

func TestRepositoryFromAPIURL(t *testing.T) {
	got := repositoryFromAPIURL("https://github.example.com/api/v3/repos/acme/orders")
	if got != "acme/orders" {
		t.Fatalf("got %q", got)
	}
}
