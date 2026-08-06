package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/wrallee/review-focus-mcp/internal/config"
	"github.com/wrallee/review-focus-mcp/internal/domain"
)

type Client struct {
	cfg  config.Config
	http *http.Client
}

func New(cfg config.Config, httpClient *http.Client) *Client {
	return &Client{cfg: cfg, http: httpClient}
}

func (c *Client) Health(ctx context.Context) (map[string]any, error) {
	var user struct {
		Login string `json:"login"`
	}
	if err := c.get(ctx, "/user", nil, &user); err != nil {
		return nil, err
	}
	out := map[string]any{"login": user.Login, "webUrl": c.cfg.GitHubURL, "apiUrl": c.cfg.GitHubAPIURL, "apiVersion": c.cfg.GitHubAPIVersion}
	if c.cfg.GitHubURL != "https://github.com" {
		var meta map[string]any
		if err := c.get(ctx, "/meta", nil, &meta); err == nil {
			if v, ok := meta["installed_version"]; ok {
				out["installedVersion"] = v
			}
		}
	}
	return out, nil
}

func (c *Client) ListReviewRequests(ctx context.Context) ([]domain.ChangeRequestSummary, error) {
	var user struct {
		Login string `json:"login"`
	}
	if err := c.get(ctx, "/user", nil, &user); err != nil {
		return nil, err
	}
	q := url.Values{"q": []string{"is:pr is:open review-requested:" + user.Login}, "sort": []string{"updated"}, "order": []string{"desc"}, "per_page": []string{"50"}}
	var resp struct {
		Items []struct {
			Number        int    `json:"number"`
			Title         string `json:"title"`
			RepositoryURL string `json:"repository_url"`
			UpdatedAt     string `json:"updated_at"`
			Draft         bool   `json:"draft"`
			HTMLURL       string `json:"html_url"`
			User          struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"items"`
	}
	if err := c.get(ctx, "/search/issues", q, &resp); err != nil {
		return nil, err
	}
	out := make([]domain.ChangeRequestSummary, 0, len(resp.Items))
	for _, it := range resp.Items {
		repo := repositoryFromAPIURL(it.RepositoryURL)
		if repo == "" {
			continue
		}
		out = append(out, domain.ChangeRequestSummary{Repository: repo, Number: it.Number, Title: it.Title, Author: it.User.Login, URL: it.HTMLURL, UpdatedAt: it.UpdatedAt, Draft: it.Draft})
	}
	return out, nil
}

func (c *Client) GetChangeRequest(ctx context.Context, repository string, number int) (domain.ChangeRequest, []domain.ChangedFile, error) {
	owner, repo, err := splitRepository(repository)
	if err != nil {
		return domain.ChangeRequest{}, nil, err
	}
	var pr pullResponse
	if err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), nil, &pr); err != nil {
		return domain.ChangeRequest{}, nil, err
	}
	cr := domain.ChangeRequest{Repository: repository, Number: number, Title: pr.Title, Body: pr.Body, Author: pr.User.Login, URL: pr.HTMLURL, BaseRef: pr.Base.Ref, HeadRef: pr.Head.Ref, BaseSHA: pr.Base.SHA, HeadSHA: pr.Head.SHA, Draft: pr.Draft, Additions: pr.Additions, Deletions: pr.Deletions, ChangedFiles: pr.ChangedFiles}
	files, err := c.listFiles(ctx, owner, repo, number)
	if err != nil {
		return domain.ChangeRequest{}, nil, err
	}
	return cr, files, nil
}

func (c *Client) SubmitReview(ctx context.Context, d domain.ReviewDraft, body string, event domain.ReviewEvent) error {
	current, _, err := c.GetChangeRequest(ctx, d.Repository, d.Number)
	if err != nil {
		return err
	}
	if current.HeadSHA != d.HeadSHA {
		return fmt.Errorf("PR head changed from %s to %s; reload before submitting", d.HeadSHA, current.HeadSHA)
	}
	owner, repo, err := splitRepository(d.Repository)
	if err != nil {
		return err
	}
	comments := make([]map[string]any, 0, len(d.Comments))
	for _, cm := range d.Comments {
		comments = append(comments, map[string]any{"path": cm.Path, "line": cm.Line, "side": cm.Side, "body": cm.Body})
	}
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", owner, repo, d.Number)
	var pending struct {
		ID int64 `json:"id"`
	}
	if err := c.post(ctx, path, map[string]any{"commit_id": d.HeadSHA, "comments": comments}, &pending); err != nil {
		return err
	}
	return c.post(ctx, path+"/"+strconv.FormatInt(pending.ID, 10)+"/events", map[string]any{"event": event, "body": body}, &map[string]any{})
}

func (c *Client) listFiles(ctx context.Context, owner, repo string, number int) ([]domain.ChangedFile, error) {
	var out []domain.ChangedFile
	for page := 1; ; page++ {
		var items []struct {
			Filename, Status, Patch, BlobURL string
			Additions, Deletions, Changes    int
		}
		q := url.Values{"per_page": []string{"100"}, "page": []string{strconv.Itoa(page)}}
		if err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d/files", owner, repo, number), q, &items); err != nil {
			return nil, err
		}
		for _, f := range items {
			out = append(out, domain.ChangedFile{Path: f.Filename, Status: f.Status, Additions: f.Additions, Deletions: f.Deletions, Changes: f.Changes, Patch: f.Patch, BlobURL: f.BlobURL})
		}
		if len(items) < 100 {
			break
		}
	}
	return out, nil
}

type pullResponse struct {
	Title        string `json:"title"`
	Body         string `json:"body"`
	HTMLURL      string `json:"html_url"`
	Draft        bool   `json:"draft"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	ChangedFiles int    `json:"changed_files"`
	User         struct {
		Login string `json:"login"`
	} `json:"user"`
	Head struct {
		SHA string `json:"sha"`
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		SHA string `json:"sha"`
		Ref string `json:"ref"`
	} `json:"base"`
}

func repositoryFromAPIURL(v string) string {
	marker := "/repos/"
	idx := strings.Index(v, marker)
	if idx < 0 {
		return ""
	}
	return strings.Trim(v[idx+len(marker):], "/")
}
func splitRepository(v string) (string, string, error) {
	parts := strings.Split(v, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("repository must be owner/name, got %q", v)
	}
	return parts[0], parts[1], nil
}
func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, out)
}
func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	return c.do(ctx, http.MethodPost, path, nil, body, out)
}
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	u := c.cfg.GitHubAPIURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", c.cfg.GitHubAPIVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		return fmt.Errorf("github %s %s: %s: %s", method, path, resp.Status, strings.TrimSpace(string(b)))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
