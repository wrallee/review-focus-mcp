package mcpapp

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrallee/review-focus-mcp/internal/analyzer"
	"github.com/wrallee/review-focus-mcp/internal/domain"
	"github.com/wrallee/review-focus-mcp/internal/provider"
	"github.com/wrallee/review-focus-mcp/internal/storage/local"
)

const resourceURI = "ui://review-focus/main"
const resourceMIME = "text/html;profile=mcp-app"

type Service struct {
	SCM      provider.SCM
	Analyzer analyzer.Analyzer
	Store    *local.Store
}
type emptyInput struct{}
type refInput struct {
	Repository string `json:"repository"`
	Number     int    `json:"number"`
}
type saveDraftInput struct {
	Repository string                 `json:"repository"`
	Number     int                    `json:"number"`
	HeadSHA    string                 `json:"headSha"`
	Comments   []domain.ReviewComment `json:"comments"`
}
type submitInput struct {
	Repository string                 `json:"repository"`
	Number     int                    `json:"number"`
	HeadSHA    string                 `json:"headSha"`
	Event      domain.ReviewEvent     `json:"event"`
	Body       string                 `json:"body,omitempty"`
	Comments   []domain.ReviewComment `json:"comments"`
}
type openOutput struct {
	ReviewRequests []domain.ChangeRequestSummary `json:"reviewRequests"`
}
type detailOutput struct {
	ChangeRequest domain.ChangeRequest `json:"changeRequest"`
	Files         []domain.ChangedFile `json:"files"`
	Analysis      *domain.Analysis     `json:"analysis,omitempty"`
	Draft         *domain.ReviewDraft  `json:"draft,omitempty"`
	DraftStale    bool                 `json:"draftStale"`
}
type okOutput struct {
	OK bool `json:"ok"`
}

func New(svc Service) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "review-focus", Title: "Review Focus", Version: "0.1.0"}, nil)
	s.AddResource(&mcp.Resource{URI: resourceURI, Name: "Review Focus", MIMEType: resourceMIME}, func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: resourceURI, MIMEType: resourceMIME, Text: appHTML}}}, nil
	})
	appMeta := mcp.Meta{"ui": map[string]any{"resourceUri": resourceURI}}
	mcp.AddTool(s, &mcp.Tool{Name: "review_focus_open", Description: "Open Review Focus and list PRs currently requesting the user's review.", Meta: appMeta}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, openOutput, error) {
		prs, err := svc.SCM.ListReviewRequests(ctx)
		return nil, openOutput{ReviewRequests: prs}, err
	})
	mcp.AddTool(s, &mcp.Tool{Name: "review_focus_get", Description: "Load one pull request, its diff, attention analysis, and pending human-review draft."}, func(ctx context.Context, _ *mcp.CallToolRequest, in refInput) (*mcp.CallToolResult, detailOutput, error) {
		return getDetail(ctx, svc, in.Repository, in.Number)
	})
	mcp.AddTool(s, &mcp.Tool{Name: "review_focus_analyze", Description: "Classify changed files into SKIP, REVIEW, or CRITICAL attention buckets."}, func(ctx context.Context, _ *mcp.CallToolRequest, in refInput) (*mcp.CallToolResult, detailOutput, error) {
		cr, files, err := svc.SCM.GetChangeRequest(ctx, in.Repository, in.Number)
		if err != nil {
			return nil, detailOutput{}, err
		}
		a, err := svc.Analyzer.Analyze(ctx, cr, files)
		if err != nil {
			return nil, detailOutput{}, err
		}
		if err := svc.Store.SaveAnalysis(a); err != nil {
			return nil, detailOutput{}, err
		}
		return getDetail(ctx, svc, in.Repository, in.Number)
	})
	mcp.AddTool(s, &mcp.Tool{Name: "review_focus_save_draft", Description: "Save the local pending human-review draft without writing to GitHub."}, func(_ context.Context, _ *mcp.CallToolRequest, in saveDraftInput) (*mcp.CallToolResult, okOutput, error) {
		err := svc.Store.SaveDraft(domain.ReviewDraft{Repository: in.Repository, Number: in.Number, HeadSHA: in.HeadSHA, Comments: in.Comments})
		return nil, okOutput{OK: err == nil}, err
	})
	mcp.AddTool(s, &mcp.Tool{Name: "review_focus_submit", Description: "Submit the pending human review after verifying the PR head SHA."}, func(ctx context.Context, _ *mcp.CallToolRequest, in submitInput) (*mcp.CallToolResult, okOutput, error) {
		switch in.Event {
		case domain.ReviewEventComment, domain.ReviewEventApprove, domain.ReviewEventRequestChanges:
		default:
			return nil, okOutput{}, fmt.Errorf("unsupported review event %q", in.Event)
		}
		d := domain.ReviewDraft{Repository: in.Repository, Number: in.Number, HeadSHA: in.HeadSHA, Comments: in.Comments}
		if err := svc.SCM.SubmitReview(ctx, d, in.Body, in.Event); err != nil {
			return nil, okOutput{}, err
		}
		return nil, okOutput{OK: true}, nil
	})
	mcp.AddTool(s, &mcp.Tool{Name: "review_focus_health", Description: "Check GitHub connectivity, authenticated user, and GHES metadata."}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, map[string]any, error) {
		out, err := svc.SCM.Health(ctx)
		return nil, out, err
	})
	return s
}

func getDetail(ctx context.Context, svc Service, repository string, number int) (*mcp.CallToolResult, detailOutput, error) {
	cr, files, err := svc.SCM.GetChangeRequest(ctx, repository, number)
	if err != nil {
		return nil, detailOutput{}, err
	}
	a, _ := svc.Store.LoadAnalysis(repository, number, cr.HeadSHA)
	d, err := svc.Store.LoadDraft(repository, number, cr.HeadSHA)
	stale := false
	if errors.Is(err, os.ErrNotExist) {
		if older, olderErr := svc.Store.FindLatestDraft(repository, number); olderErr == nil && older.HeadSHA != cr.HeadSHA {
			stale = true
		}
	}
	return nil, detailOutput{ChangeRequest: cr, Files: files, Analysis: a, Draft: d, DraftStale: stale}, nil
}

func RunStdio(ctx context.Context, server *mcp.Server) error {
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
func EnsureEmbeddedUI() error {
	if appHTML == "" || appHTML == "<!-- built UI placeholder; run make ui -->\n" {
		return errors.New("MCP App UI is not built; run `make build`")
	}
	return nil
}
