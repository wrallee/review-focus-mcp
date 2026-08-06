package provider

import (
	"context"
	"github.com/wrallee/review-focus-mcp/internal/domain"
)

type SCM interface {
	Health(context.Context) (map[string]any, error)
	ListReviewRequests(context.Context) ([]domain.ChangeRequestSummary, error)
	GetChangeRequest(context.Context, string, int) (domain.ChangeRequest, []domain.ChangedFile, error)
	SubmitReview(context.Context, domain.ReviewDraft, string, domain.ReviewEvent) error
}
