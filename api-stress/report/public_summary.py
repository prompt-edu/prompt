"""Write a public summary of one run, with P0 findings reduced to a count.

CI output on a public repository is readable by anyone, and a P0 finding (auth
bypass, cross-course IDOR, a service going down) can describe an unfixed
vulnerability. So this summary lists P1/P2 findings in full but only counts P0
ones; their details stay in the run's own report.md, reproducible locally.

Usage: public_summary.py <run_dir> <out_dir>
Reads:  <run_dir>/findings.json, <run_dir>/meta.json
Writes: <out_dir>/summary.md, <out_dir>/findings.json (P0 entries removed)
"""

from __future__ import annotations

import json
import pathlib
import sys

REDACTED = "P0"
# Long enough to read the status histogram, short enough to keep the table narrow.
OBSERVED_MAX = 120
LOCAL_RUN = "uv run --project api-stress api-stress-test"


def redact(findings: list[dict]) -> tuple[list[dict], int]:
    """Split off the P0 findings: return the publishable rest and the P0 count."""
    public = [f for f in findings if f.get("priority") != REDACTED]
    return public, len(findings) - len(public)


def cell(value) -> str:
    """Render a value as one Markdown table cell (no pipes or line breaks)."""
    text = "" if value is None else str(value)
    return text.replace("|", "\\|").replace("\n", " ")


def findings_table(findings: list[dict]) -> list[str]:
    lines = [
        "| Kind | Method | Endpoint | Title | Observed |",
        "|------|--------|----------|-------|----------|",
    ]
    for f in findings:
        observed = (f.get("observed") or "")[:OBSERVED_MAX]
        lines.append(
            f"| {cell(f.get('kind'))} | {cell(f.get('method'))} | `{cell(f.get('endpoint'))}` "
            f"| {cell(f.get('title'))} | {cell(observed)} |"
        )
    return lines


def render(findings: list[dict], meta: dict) -> str:
    public, p0 = redact(findings)
    by_priority = {
        pri: [f for f in public if f.get("priority") == pri] for pri in ("P1", "P2")
    }

    lines = [
        "## API stress test",
        "",
        f"Intensity `{meta.get('intensity', '?')}`, scenarios `{meta.get('scenarios', '?')}`, "
        f"{meta.get('catalog_count', '?')} cataloged endpoints.",
        "",
        "| Priority | Findings |",
        "|----------|----------|",
        f"| P0 (auth bypass / IDOR / outage) | {p0} |",
        f"| P1 (5xx on valid request / degradation / DoS) | {len(by_priority['P1'])} |",
        f"| P2 (input robustness / slow) | {len(by_priority['P2'])} |",
        "",
    ]
    if p0:
        lines += [
            "> [!CAUTION]",
            f"> {p0} P0 finding(s) are not listed here: this output is public and they may "
            f"describe an unfixed vulnerability. Reproduce locally with `{LOCAL_RUN}` "
            "(see `api-stress/README.md`) and read the run's `report.md`.",
            "",
        ]
    for pri, label in (("P1", "P1 - High"), ("P2", "P2 - Robustness / Performance")):
        group = by_priority[pri]
        if not group:
            continue
        lines += [
            "<details>",
            f"<summary>{label} ({len(group)})</summary>",
            "",
            *findings_table(group),
            "",
            "</details>",
            "",
        ]
    return "\n".join(lines)


def main() -> int:
    if len(sys.argv) != 3:
        print("usage: public_summary.py <run_dir> <out_dir>", file=sys.stderr)
        return 2
    run_dir, out_dir = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
    findings = json.loads((run_dir / "findings.json").read_text())
    meta_path = run_dir / "meta.json"
    meta = json.loads(meta_path.read_text()) if meta_path.exists() else {}

    out_dir.mkdir(parents=True, exist_ok=True)
    public, p0 = redact(findings)
    (out_dir / "findings.json").write_text(json.dumps(public, indent=2))
    (out_dir / "summary.md").write_text(render(findings, meta))
    print(f"public summary -> {out_dir/'summary.md'} ({p0} P0 finding(s) redacted)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
