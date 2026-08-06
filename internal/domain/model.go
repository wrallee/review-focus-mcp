package domain

type Attention string

const (
	AttentionSkip     Attention = "SKIP"
	AttentionReview   Attention = "REVIEW"
	AttentionCritical Attention = "CRITICAL"
)

type ChangeRequestSummary struct {
	Repository string `json:"repository"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	URL        string `json:"url"`
	UpdatedAt  string `json:"updatedAt"`
	Draft      bool   `json:"draft"`
}

type ChangeRequest struct {
	Repository   string `json:"repository"`
	Number       int    `json:"number"`
	Title        string `json:"title"`
	Body         string `json:"body,omitempty"`
	Author       string `json:"author"`
	URL          string `json:"url"`
	BaseRef      string `json:"baseRef"`
	HeadRef      string `json:"headRef"`
	BaseSHA      string `json:"baseSha"`
	HeadSHA      string `json:"headSha"`
	Draft        bool   `json:"draft"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	ChangedFiles int    `json:"changedFiles"`
}

type ChangedFile struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
	Patch     string `json:"patch,omitempty"`
	BlobURL   string `json:"blobUrl,omitempty"`
}

type FileAnalysis struct {
	Path            string    `json:"path"`
	Attention       Attention `json:"attention"`
	ChangeTypes     []string  `json:"changeTypes"`
	Reason          string    `json:"reason"`
	Explanation     string    `json:"explanation,omitempty"`
	PotentialImpact string    `json:"potentialImpact,omitempty"`
	ReviewPoints    []string  `json:"reviewPoints,omitempty"`
	Confidence      float64   `json:"confidence"`
}

type Analysis struct {
	Provider   string         `json:"provider"`
	Repository string         `json:"repository"`
	Number     int            `json:"number"`
	HeadSHA    string         `json:"headSha"`
	Analyzer   string         `json:"analyzer"`
	Files      []FileAnalysis `json:"files"`
}

type ReviewComment struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Side string `json:"side"`
	Body string `json:"body"`
}

type ReviewDraft struct {
	Repository string          `json:"repository"`
	Number     int             `json:"number"`
	HeadSHA    string          `json:"headSha"`
	Comments   []ReviewComment `json:"comments"`
}

type ReviewEvent string

const (
	ReviewEventComment        ReviewEvent = "COMMENT"
	ReviewEventApprove        ReviewEvent = "APPROVE"
	ReviewEventRequestChanges ReviewEvent = "REQUEST_CHANGES"
)
