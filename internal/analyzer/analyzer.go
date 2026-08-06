package analyzer

import (
	"context"
	"strings"

	"github.com/wrallee/review-focus-mcp/internal/domain"
)

type Analyzer interface {
	Name() string
	Analyze(context.Context, domain.ChangeRequest, []domain.ChangedFile) (domain.Analysis, error)
}

type Rules struct{}

func (Rules) Name() string { return "rules-v0" }

func (r Rules) Analyze(_ context.Context, cr domain.ChangeRequest, files []domain.ChangedFile) (domain.Analysis, error) {
	out := domain.Analysis{Provider: "github", Repository: cr.Repository, Number: cr.Number, HeadSHA: cr.HeadSHA, Analyzer: r.Name(), Files: make([]domain.FileAnalysis, 0, len(files))}
	for _, f := range files {
		path := strings.ToLower(f.Path)
		patch := strings.ToLower(f.Patch)
		a := domain.FileAnalysis{Path: f.Path, Attention: domain.AttentionReview, Reason: "Implementation behavior changed", Explanation: "This change is not obviously mechanical, so it remains in the normal human-review path.", ChangeTypes: []string{"IMPLEMENTATION"}, Confidence: 0.65}
		switch {
		case strings.HasSuffix(path, "_test.go"), strings.Contains(path, "/test/"), strings.Contains(path, "/tests/"), strings.HasSuffix(path, ".md"):
			a.Attention = domain.AttentionSkip
			a.Reason = "Test or documentation-only change"
			a.Explanation = "The change is useful context but is folded by default to reduce review noise."
			a.ChangeTypes = []string{"TEST_OR_DOC"}
			a.Confidence = 0.9
		case strings.Contains(path, "migration"), strings.Contains(path, "schema"), strings.Contains(patch, "transaction"), strings.Contains(patch, "authorize"), strings.Contains(patch, "permission"), strings.Contains(patch, "mutex"), strings.Contains(patch, "lock("), strings.Contains(patch, "cache"):
			a.Attention = domain.AttentionCritical
			a.Reason = "High-impact state or control-flow concern changed"
			a.Explanation = "This file touches persistence, transaction, authorization, concurrency, or cache behavior where small changes can have system-wide effects."
			a.PotentialImpact = "Consistency, access control, concurrency, cache correctness, or production data flow can change."
			a.ReviewPoints = []string{"Verify behavioral invariants", "Verify rollback/failure behavior", "Check compatibility with existing callers"}
			a.ChangeTypes = []string{"HIGH_IMPACT"}
			a.Confidence = 0.8
		}
		out.Files = append(out.Files, a)
	}
	return out, nil
}
