package analyzer

import (
	"context"
	"github.com/wrallee/review-focus-mcp/internal/domain"
	"testing"
)

func TestRulesClassify(t *testing.T) {
	cr := domain.ChangeRequest{Repository: "o/r", Number: 1, HeadSHA: "abc"}
	files := []domain.ChangedFile{{Path: "README.md"}, {Path: "internal/cache/store.go", Patch: "+ cache.Set(key)"}, {Path: "service/user.go", Patch: "+ return user"}}
	a, err := (Rules{}).Analyze(context.Background(), cr, files)
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Attention{domain.AttentionSkip, domain.AttentionCritical, domain.AttentionReview}
	for i := range want {
		if a.Files[i].Attention != want[i] {
			t.Fatalf("file %d got %s want %s", i, a.Files[i].Attention, want[i])
		}
	}
}
