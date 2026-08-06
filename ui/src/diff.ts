export interface DiffLine { text: string; kind: "add" | "del" | "ctx" | "meta"; oldLine?: number; newLine?: number; side?: "LEFT" | "RIGHT"; line?: number; }

export function parsePatch(patch = ""): DiffLine[] {
  let oldLine = 0;
  let newLine = 0;
  const rows: DiffLine[] = [];
  for (const raw of patch.split("\n")) {
    const match = raw.match(/^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/);
    if (match) {
      oldLine = Number(match[1]); newLine = Number(match[2]); rows.push({ text: raw, kind: "meta" }); continue;
    }
    if (raw.startsWith("+") && !raw.startsWith("+++")) { rows.push({ text: raw, kind: "add", newLine, side: "RIGHT", line: newLine }); newLine++; continue; }
    if (raw.startsWith("-") && !raw.startsWith("---")) { rows.push({ text: raw, kind: "del", oldLine, side: "LEFT", line: oldLine }); oldLine++; continue; }
    rows.push({ text: raw, kind: "ctx", oldLine, newLine, side: "RIGHT", line: newLine }); oldLine++; newLine++;
  }
  return rows;
}
