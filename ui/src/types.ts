export type Attention = "SKIP" | "REVIEW" | "CRITICAL";

export interface ChangeRequestSummary { repository: string; number: number; title: string; author: string; url: string; updatedAt: string; draft: boolean; }
export interface ChangeRequest { repository: string; number: number; title: string; body?: string; author: string; url: string; baseRef: string; headRef: string; baseSha: string; headSha: string; draft: boolean; additions: number; deletions: number; changedFiles: number; }
export interface ChangedFile { path: string; status: string; additions: number; deletions: number; changes: number; patch?: string; blobUrl?: string; }
export interface FileAnalysis { path: string; attention: Attention; changeTypes: string[]; reason: string; explanation?: string; potentialImpact?: string; reviewPoints?: string[]; confidence: number; }
export interface Analysis { provider: string; repository: string; number: number; headSha: string; analyzer: string; files: FileAnalysis[]; }
export interface ReviewComment { path: string; line: number; side: "LEFT" | "RIGHT"; body: string; }
export interface DraftReview { repository: string; number: number; headSha: string; comments: ReviewComment[]; }
export interface DetailOutput { changeRequest: ChangeRequest; files: ChangedFile[]; analysis?: Analysis; draft?: DraftReview; draftStale: boolean; }
export interface OpenOutput { reviewRequests: ChangeRequestSummary[]; }
