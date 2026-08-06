import { useEffect, useMemo, useState } from "react";
import { callTool, subscribeToolResults } from "./mcp";
import { parsePatch } from "./diff";
import type { Attention, ChangeRequestSummary, DetailOutput, FileAnalysis, OpenOutput, ReviewComment } from "./types";
import "./styles.css";

const order: Attention[] = ["CRITICAL", "REVIEW", "SKIP"];

export default function ReviewFocusApp() {
  const [requests, setRequests] = useState<ChangeRequestSummary[]>([]);
  const [detail, setDetail] = useState<DetailOutput>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [comments, setComments] = useState<ReviewComment[]>([]);
  const [summary, setSummary] = useState("");

  useEffect(() => {
    const unsubscribe = subscribeToolResults((result) => {
      try {
        const data = result.structuredContent as OpenOutput | DetailOutput | undefined;
        if (data && "reviewRequests" in data) setRequests(data.reviewRequests);
        if (data && "changeRequest" in data) hydrateDetail(data);
      } catch { /* host may deliver unrelated tool results */ }
    });
    void refreshInbox();
    return unsubscribe;
  }, []);

  function hydrateDetail(data: DetailOutput) {
    setDetail(data);
    setComments(data.draft?.comments ?? []);
  }

  async function refreshInbox() {
    setLoading(true); setError("");
    try { setRequests((await callTool<OpenOutput>("review_focus_open")).reviewRequests); }
    catch (e) { setError(message(e)); }
    finally { setLoading(false); }
  }

  async function selectPR(pr: ChangeRequestSummary) {
    setLoading(true); setError("");
    try { hydrateDetail(await callTool<DetailOutput>("review_focus_get", { repository: pr.repository, number: pr.number })); }
    catch (e) { setError(message(e)); }
    finally { setLoading(false); }
  }

  async function analyze() {
    if (!detail) return;
    setLoading(true); setError("");
    try { hydrateDetail(await callTool<DetailOutput>("review_focus_analyze", { repository: detail.changeRequest.repository, number: detail.changeRequest.number })); }
    catch (e) { setError(message(e)); }
    finally { setLoading(false); }
  }

  async function saveDraft(next: ReviewComment[]) {
    if (!detail) return;
    setComments(next);
    await callTool("review_focus_save_draft", { repository: detail.changeRequest.repository, number: detail.changeRequest.number, headSha: detail.changeRequest.headSha, comments: next });
  }

  async function submit(event: "COMMENT" | "APPROVE" | "REQUEST_CHANGES") {
    if (!detail) return;
    if (event === "REQUEST_CHANGES" && !summary.trim()) { setError("Request changes requires a review summary."); return; }
    setLoading(true); setError("");
    try {
      await callTool("review_focus_submit", { repository: detail.changeRequest.repository, number: detail.changeRequest.number, headSha: detail.changeRequest.headSha, event, body: summary, comments });
      setComments([]); setSummary("");
      await refreshInbox();
    } catch (e) { setError(message(e)); }
    finally { setLoading(false); }
  }

  const analysisMap = useMemo(() => new Map(detail?.analysis?.files.map((x) => [x.path, x]) ?? []), [detail?.analysis]);

  return <div className="shell">
    <header>
      <div><strong>Review Focus</strong><span>human review, less noise</span></div>
      <button className="ghost" onClick={refreshInbox}>Refresh</button>
    </header>
    {error && <div className="error">{error}</div>}
    <div className="layout">
      <aside>
        <div className="aside-title">Review requested <span>{requests.length}</span></div>
        {requests.map((pr) => <button key={`${pr.repository}#${pr.number}`} className={`pr ${detail?.changeRequest.repository === pr.repository && detail?.changeRequest.number === pr.number ? "active" : ""}`} onClick={() => selectPR(pr)}>
          <small>{pr.repository} · #{pr.number}</small><b>{pr.title}</b><span>@{pr.author}</span>
        </button>)}
        {!loading && requests.length === 0 && <p className="muted">No review requests.</p>}
      </aside>
      <main>
        {!detail ? <div className="empty">Select a pull request to start focused review.</div> : <>
          <section className="pr-head">
            <div><small>{detail.changeRequest.repository} · #{detail.changeRequest.number}</small><h1>{detail.changeRequest.title}</h1><p>{detail.changeRequest.baseRef} ← {detail.changeRequest.headRef} · <span className="plus">+{detail.changeRequest.additions}</span> <span className="minus">-{detail.changeRequest.deletions}</span> · {detail.changeRequest.changedFiles} files</p></div>
            <div className="actions"><button className="primary" onClick={analyze}>{detail.analysis ? "Re-run analysis" : "Analyze changes"}</button><a href={detail.changeRequest.url} target="_blank" rel="noreferrer">Open SCM ↗</a></div>
          </section>
          {detail.draftStale && <div className="warning">Saved draft belongs to an older HEAD. Reloaded without those comments.</div>}
          {!detail.analysis ? <div className="analysis-empty"><b>Analysis not run</b><p>Run classification to fold low-value changes and surface review-critical changes.</p></div> : <>
            <AttentionSummary files={detail.analysis.files}/>
            {order.map((attention) => <ReviewGroup key={attention} attention={attention} detail={detail} analyses={analysisMap} comments={comments} onComments={saveDraft}/>) }
          </>}
          <section className="submit-bar">
            <div><b>{comments.length} pending comment{comments.length === 1 ? "" : "s"}</b><textarea value={summary} onChange={(e) => setSummary(e.target.value)} placeholder="Review summary (required for request changes)"/></div>
            <div className="submit-actions"><button onClick={() => submit("COMMENT")}>Comment</button><button onClick={() => submit("REQUEST_CHANGES")}>Request changes</button><button className="approve" onClick={() => submit("APPROVE")}>Approve</button></div>
          </section>
        </>}
      </main>
    </div>
    {loading && <div className="loading">Working…</div>}
  </div>;
}

function AttentionSummary({ files }: { files: FileAnalysis[] }) {
  const counts = Object.fromEntries(order.map((a) => [a, files.filter((f) => f.attention === a).length]));
  return <section className="summary">{order.map((a) => <div key={a} className={`metric ${a.toLowerCase()}`}><span>{a}</span><b>{counts[a]}</b></div>)}</section>;
}

function ReviewGroup({ attention, detail, analyses, comments, onComments }: { attention: Attention; detail: DetailOutput; analyses: Map<string, FileAnalysis>; comments: ReviewComment[]; onComments: (x: ReviewComment[]) => Promise<void> }) {
  const files = detail.files.filter((file) => analyses.get(file.path)?.attention === attention);
  const [open, setOpen] = useState(attention !== "SKIP");
  if (files.length === 0) return null;
  return <section className={`group ${attention.toLowerCase()}`}>
    <button className="group-title" onClick={() => setOpen(!open)}><span>{open ? "▾" : "▸"} {attention}</span><b>{files.length} files</b></button>
    {open && files.map((file) => <FileDiff key={file.path} file={file} analysis={analyses.get(file.path)!} comments={comments} onComments={onComments}/>) }
  </section>;
}

function FileDiff({ file, analysis, comments, onComments }: { file: DetailOutput["files"][number]; analysis: FileAnalysis; comments: ReviewComment[]; onComments: (x: ReviewComment[]) => Promise<void> }) {
  const [open, setOpen] = useState(analysis.attention !== "SKIP");
  const [draftLine, setDraftLine] = useState<{line:number; side:"LEFT"|"RIGHT"} | null>(null);
  const [body, setBody] = useState("");
  const rows = useMemo(() => parsePatch(file.patch), [file.patch]);
  async function addComment() {
    if (!draftLine || !body.trim()) return;
    await onComments([...comments, { path: file.path, line: draftLine.line, side: draftLine.side, body: body.trim() }]);
    setBody(""); setDraftLine(null);
  }
  return <article className="file">
    <button className="file-head" onClick={() => setOpen(!open)}><span>{open ? "▾" : "▸"} <code>{file.path}</code></span><span><em>+{file.additions}</em> <del>-{file.deletions}</del></span></button>
    <div className="why"><span>{analysis.changeTypes.join(" · ")}</span><b>{analysis.reason}</b>{analysis.attention === "CRITICAL" && <div className="critical-note"><p>{analysis.explanation}</p><p><strong>Impact:</strong> {analysis.potentialImpact}</p>{analysis.reviewPoints?.length ? <ul>{analysis.reviewPoints.map((p) => <li key={p}>{p}</li>)}</ul> : null}</div>}</div>
    {open && <div className="diff">
      {rows.length === 0 ? <div className="no-patch">Patch unavailable (binary or GitHub omitted the patch).</div> : rows.map((row, i) => <div key={i} className={`diff-row ${row.kind}`}>
        <span className="comment-slot">{row.line && row.side ? <button title="Add review comment" onClick={() => setDraftLine({line: row.line!, side: row.side!})}>+</button> : null}</span>
        <span className="ln">{row.oldLine ?? ""}</span><span className="ln">{row.newLine ?? ""}</span><code>{row.text || " "}</code>
      </div>)}
      {draftLine && <div className="comment-editor"><b>Comment on {draftLine.side === "RIGHT" ? "new" : "old"} line {draftLine.line}</b><textarea autoFocus value={body} onChange={(e) => setBody(e.target.value)} placeholder="Leave a review comment…"/><div><button onClick={() => setDraftLine(null)}>Cancel</button><button className="primary" onClick={addComment}>Add to review</button></div></div>}
      {comments.filter((c) => c.path === file.path).map((c, i) => <div className="pending" key={`${c.line}-${i}`}><span>Pending · {c.side}:{c.line}</span><p>{c.body}</p></div>)}
    </div>}
  </article>;
}

function message(e: unknown) { return e instanceof Error ? e.message : String(e); }
