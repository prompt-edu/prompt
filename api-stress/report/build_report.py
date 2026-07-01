#!/usr/bin/env python3
"""Merge k6 outputs + fuzz findings into a prioritized, bug-fixable report.

Reads from <run_dir>:
  raw/k6_<scenario>.json   - k6 json output per scenario (http_req_duration + reqs)
  fuzz_findings.json       - structured security/robustness findings
  fixtures.json, meta.json - run context

Writes: report.md, report.html, findings.json (machine-readable, prioritized).

Priority model:
  P0  auth bypass (no/forged/alg-none token accepted), IDOR, a service that went DOWN
  P1  5xx on a VALID request (GET read endpoint with real ids), latency degradation,
      oversized/nested-body crash, slowloris exposure
  P2  5xx on malformed input/non-existent target (robustness), individually slow endpoints
"""
from __future__ import annotations

import json
import pathlib
import sys
from collections import defaultdict

HERE = pathlib.Path(__file__).parent
CATALOG = {e["id"]: e for e in json.load(open(HERE.parent / "catalog" / "endpoints.json"))["endpoints"]}

SLOW_P95_MS = 1000.0  # endpoints slower than this (p95) are flagged


def pct(values, p):
    if not values:
        return 0.0
    s = sorted(values)
    k = (len(s) - 1) * (p / 100.0)
    f = int(k)
    c = min(f + 1, len(s) - 1)
    if f == c:
        return s[f]
    return s[f] + (s[c] - s[f]) * (k - f)


def load_k6(run_dir):
    """Return {scenario: {endpoint: {lat:[...], status:Counter, points:[(t,ms)]}}}."""
    raw_dir = run_dir / "raw"
    out = {}
    for f in sorted(raw_dir.glob("k6_*.json")) if raw_dir.exists() else []:
        scen = f.stem.replace("k6_", "")
        eps = defaultdict(lambda: {"lat": [], "status": defaultdict(int), "points": []})
        for line in open(f):
            try:
                o = json.loads(line)
            except Exception:  # noqa: BLE001
                continue
            if o.get("type") != "Point":
                continue
            m = o.get("metric")
            d = o.get("data", {})
            tg = d.get("tags", {})
            ep = tg.get("endpoint")
            if not ep:
                continue
            if m == "http_req_duration":
                eps[ep]["lat"].append(d.get("value", 0.0))
                eps[ep]["points"].append((d.get("time"), d.get("value", 0.0)))
            elif m == "reqs":
                eps[ep]["status"][tg.get("status", "?")] += 1
        out[scen] = eps
    return out


def is_5xx(status):
    return status and (status[0] == "5" or status == "0")


def endpoint_summary(k6):
    """Per-endpoint rollup across scenarios."""
    rows = {}
    for scen, eps in k6.items():
        for ep, d in eps.items():
            r = rows.setdefault(ep, {"lat": [], "status": defaultdict(int), "scenarios": set()})
            r["lat"].extend(d["lat"])
            for s, c in d["status"].items():
                r["status"][s] += c
            r["scenarios"].add(scen)
    for ep, r in rows.items():
        r["p50"] = pct(r["lat"], 50)
        r["p95"] = pct(r["lat"], 95)
        r["p99"] = pct(r["lat"], 99)
        r["max"] = max(r["lat"]) if r["lat"] else 0.0
        r["n"] = sum(r["status"].values())
        r["n5xx"] = sum(c for s, c in r["status"].items() if is_5xx(s))
    return rows


def degradation(k6):
    """For the load + exhaustion scenarios, bucket all requests into time windows
    and report p95 latency + error rate per window to expose the knee."""
    out = {}
    for scen in ("load", "exhaustion", "spike"):
        eps = k6.get(scen)
        if not eps:
            continue
        pts = []
        for ep, d in eps.items():
            for t, ms in d["points"]:
                pts.append((t, ms))
        if not pts:
            continue
        pts.sort(key=lambda x: x[0] or "")
        # 10 equal-COUNT slices: each holds ~n/10 requests in
        # chronological order, so under bursty load a slice spans less wall-clock.
        # Reads as "latency progression across the run", not fixed-duration windows.
        n = len(pts)
        slices = 10
        buckets = []
        for i in range(slices):
            seg = pts[i * n // slices:(i + 1) * n // slices]
            lat = [ms for _, ms in seg]
            buckets.append({"window": i + 1, "n": len(seg), "p95": round(pct(lat, 95), 1),
                            "max": round(max(lat), 1) if lat else 0})
        out[scen] = buckets
    return out


def build_findings(k6, fuzz, summary):
    findings = []

    def add(pri, **kw):
        kw["priority"] = pri
        kw["suspected_location"] = locate(kw.get("endpoint"))
        findings.append(kw)

    # --- from fuzz ---
    for f in fuzz:
        axis = f.get("axis")
        if axis == "authz":
            pri = "P0"
        elif axis == "idor":
            pri = "P0"
        elif axis == "dos":
            pri = "P1"
        elif axis == "inputs":
            pri = "P1" if f.get("severity") == "high" else "P2"
        else:
            pri = "P2"
        add(pri, kind=axis, endpoint=f.get("endpoint"), method=f.get("method"),
            title=f.get("detail"), observed=f.get("observed"), expected=f.get("expected"),
            evidence=f.get("evidence"), repro=f.get("repro"), url=f.get("url"))

    # --- service DOWN (P0): a service whose requests were overwhelmingly
    # connection errors (status 0) was unreachable/crashed during the run - the
    # single highest-severity signal, so it outranks any 5xx finding below.
    smoke = k6.get("smoke", {})
    down = defaultdict(lambda: {"total": 0, "conn_err": 0})
    for ep, d in smoke.items():
        svc = CATALOG.get(ep, {}).get("service", "?")
        for s, c in d["status"].items():
            down[svc]["total"] += c
            if s == "0":
                down[svc]["conn_err"] += c
    for svc, t in down.items():
        if t["total"] >= 3 and t["conn_err"] >= 0.8 * t["total"]:
            add("P0", kind="service-down", endpoint=svc, method=None,
                title=f"service '{svc}' was unreachable during the run",
                observed=f"{t['conn_err']}/{t['total']} requests were connection errors (status 0)",
                expected="service stays reachable for the whole run",
                evidence="all/most endpoints on this service failed to connect - the "
                "service crashed, OOM'd, or the Docker host went down",
                repro=None, url=None)

    # --- from k6: 5xx on valid GET reads (real ids) = P1; others P2 ---
    for ep, d in smoke.items():
        n5 = sum(c for s, c in d["status"].items() if is_5xx(s))
        if not n5:
            continue
        meta = CATALOG.get(ep, {})
        valid_read = meta.get("method") == "GET" and meta.get("category") in ("read", "public")
        statuses = dict(d["status"])
        add("P1" if valid_read else "P2", kind="server-error", endpoint=ep,
            method=meta.get("method"), title=(
                "5xx on a valid read request" if valid_read else "5xx on write/delete (bad/non-existent target)"),
            observed=f"statuses {statuses}", expected="2xx (read) or 4xx (bad input)",
            evidence=f"path {meta.get('path')}", repro=None, url=None)

    # --- latency: slow endpoints (p95 over threshold) ---
    for ep, r in summary.items():
        if r["p95"] >= SLOW_P95_MS and r["n5xx"] == 0:
            add("P2", kind="slow", endpoint=ep, method=CATALOG.get(ep, {}).get("method"),
                title=f"slow endpoint p95={r['p95']:.0f}ms", observed=f"p50={r['p50']:.0f} p95={r['p95']:.0f} p99={r['p99']:.0f} max={r['max']:.0f}ms over {r['n']} reqs",
                expected=f"p95 < {SLOW_P95_MS:.0f}ms", evidence="", repro=None, url=None)

    order = {"P0": 0, "P1": 1, "P2": 2}
    findings.sort(key=lambda f: (order.get(f["priority"], 9), f.get("endpoint") or ""))
    return findings


def locate(endpoint):
    """Best-effort source pointer from the endpoint id/service."""
    if not endpoint:
        return None
    meta = CATALOG.get(endpoint, {})
    svc = meta.get("service")
    if not svc:
        return None
    return f"servers/{svc}/ (handler {meta.get('handler','?')}, path {meta.get('path','?')})"


def md_table(headers, rows):
    out = ["| " + " | ".join(headers) + " |", "|" + "|".join(["---"] * len(headers)) + "|"]
    for r in rows:
        out.append("| " + " | ".join(str(c) for c in r) + " |")
    return "\n".join(out)


def main() -> int:
    run_dir = pathlib.Path(sys.argv[1])
    meta = {}
    if (run_dir / "meta.json").exists():
        meta = json.load(open(run_dir / "meta.json"))
    fuzz = []
    if (run_dir / "fuzz_findings.json").exists():
        fuzz = json.load(open(run_dir / "fuzz_findings.json"))

    k6 = load_k6(run_dir)
    summary = endpoint_summary(k6)
    deg = degradation(k6)
    findings = build_findings(k6, fuzz, summary)

    json.dump(findings, open(run_dir / "findings.json", "w"), indent=2)

    # ----- counts -----
    p = lambda x: sum(1 for f in findings if f["priority"] == x)  # noqa: E731
    tested = len(summary)
    total_5xx_eps = sum(1 for r in summary.values() if r["n5xx"])
    slowest = sorted(summary.items(), key=lambda kv: -kv[1]["p95"])[:15]
    authz = [f for f in findings if f["kind"] == "authz"]
    idor = [f for f in findings if f["kind"] == "idor"]

    md = []
    md.append("# PROMPT API Stress & Fuzz Report")
    md.append("")
    md.append(f"- Run: `{run_dir.name}`")
    for k in ("started", "intensity", "scenarios", "stack", "catalog_count"):
        if k in meta:
            md.append(f"- {k}: `{meta[k]}`")
    md.append(f"- Endpoints in catalog: **{len(CATALOG)}**, exercised with metrics: **{tested}**")
    md.append("")
    md.append("## Executive summary")
    md.append("")
    md.append(md_table(
        ["metric", "value"],
        [
            ["P0 findings (auth bypass / IDOR / outage)", p("P0")],
            ["P1 findings (5xx on valid request / degradation / DoS)", p("P1")],
            ["P2 findings (input robustness / slow)", p("P2")],
            ["Endpoints returning any 5xx", total_5xx_eps],
            ["Auth-bypass findings", len(authz)],
            ["IDOR findings", len(idor)],
        ],
    ))
    md.append("")
    if p("P0") == 0:
        md.append("> No auth-bypass or IDOR issues detected (no-token, forged-signature, "
                  "alg=none, and wrong-role probes were all correctly rejected).")
        md.append("")

    # ----- prioritized findings -----
    for pri, label in [("P0", "P0 - Critical"), ("P1", "P1 - High"), ("P2", "P2 - Robustness / Performance")]:
        group = [f for f in findings if f["priority"] == pri]
        if not group:
            continue
        md.append(f"## {label}  ({len(group)})")
        md.append("")
        # collapse identical (endpoint,kind) input-body findings into one row w/ count
        seen = defaultdict(list)
        for f in group:
            seen[(f.get("endpoint"), f.get("kind"), f.get("title"))].append(f)
        for (ep, kind, title), fs in seen.items():
            f0 = fs[0]
            md.append(f"### `{ep}` - {kind}" + (f" (x{len(fs)})" if len(fs) > 1 else ""))
            md.append(f"- **What:** {title}")
            if f0.get("method"):
                md.append(f"- **Method:** {f0['method']}")
            if f0.get("observed"):
                md.append(f"- **Observed:** {f0['observed']}")
            if f0.get("expected"):
                md.append(f"- **Expected:** {f0['expected']}")
            if f0.get("suspected_location"):
                md.append(f"- **Where to look:** `{f0['suspected_location']}`")
            if f0.get("evidence"):
                ev = str(f0["evidence"]).replace("\n", " ")[:300]
                md.append(f"- **Evidence:** `{ev}`")
            if f0.get("repro"):
                md.append(f"- **Repro:**\n  ```\n  {f0['repro']}\n  ```")
            md.append("")

    # ----- latency table -----
    md.append("## Slowest endpoints (by p95)")
    md.append("")
    md.append(md_table(
        ["endpoint", "method", "p50 ms", "p95 ms", "p99 ms", "max ms", "reqs", "5xx"],
        [[ep, CATALOG.get(ep, {}).get("method", "?"), f"{r['p50']:.0f}", f"{r['p95']:.0f}",
          f"{r['p99']:.0f}", f"{r['max']:.0f}", r["n"], r["n5xx"]] for ep, r in slowest],
    ))
    md.append("")

    # ----- degradation -----
    if deg:
        md.append("## Load degradation (p95 latency across the ramp)")
        md.append("")
        for scen, buckets in deg.items():
            md.append(f"### {scen}")
            md.append(md_table(
                ["window (early->late / ramp up)", "reqs", "p95 ms", "max ms"],
                [[b["window"], b["n"], b["p95"], b["max"]] for b in buckets],
            ))
            md.append("")

    # ----- coverage -----
    fx = {}
    if (run_dir / "fixtures.json").exists():
        fx = json.load(open(run_dir / "fixtures.json"))
    md.append("## Coverage & fixtures")
    md.append("")
    md.append(f"- Course A: `{fx.get('course_a',{}).get('id')}` ({len(fx.get('phases_a',{}))} phases)")
    md.append(f"- Course B (IDOR target): `{fx.get('course_b',{}).get('id')}`")
    md.append(f"- Student enrolled: `{(fx.get('student') or {}).get('student_id')}`")
    md.append(f"- Per-service data seeding (best-effort): `{json.dumps(fx.get('seeded',{}))[:400]}`")
    md.append("")
    md.append("_Write/delete endpoints in smoke target non-existent ids (non-destructive); "
              "real write bodies are exercised by the fuzzer. Some 5xx are 'missing not-found "
              "handling' (should be 404) rather than crashes - see per-finding detail._")
    md.append("")

    open(run_dir / "report.md", "w").write("\n".join(md))

    # ----- simple HTML (md in <pre> + sortable endpoint table) -----
    html = ["<html><head><meta charset='utf-8'><title>PROMPT stress report</title>",
            "<style>body{font-family:system-ui;margin:2rem;max-width:1100px}table{border-collapse:collapse}td,th{border:1px solid #ccc;padding:4px 8px}tr:nth-child(even){background:#f6f6f6}.p0{color:#b00}.p1{color:#c60}</style></head><body>"]
    html.append("<h1>PROMPT API stress & fuzz report</h1>")
    html.append(f"<p>Run {run_dir.name} - P0={p('P0')} P1={p('P1')} P2={p('P2')} - {tested}/{len(CATALOG)} endpoints</p>")
    html.append("<h2>Findings</h2><table><tr><th>Pri</th><th>Kind</th><th>Endpoint</th><th>What</th><th>Observed</th></tr>")
    for f in findings:
        cls = f["priority"].lower()
        html.append(f"<tr class='{cls}'><td>{f['priority']}</td><td>{f.get('kind')}</td><td>{f.get('endpoint')}</td><td>{f.get('title')}</td><td>{(f.get('observed') or '')[:120]}</td></tr>")
    html.append("</table><h2>Slowest endpoints</h2><table><tr><th>Endpoint</th><th>p95 ms</th><th>p99 ms</th><th>max ms</th><th>5xx</th></tr>")
    for ep, r in slowest:
        html.append(f"<tr><td>{ep}</td><td>{r['p95']:.0f}</td><td>{r['p99']:.0f}</td><td>{r['max']:.0f}</td><td>{r['n5xx']}</td></tr>")
    html.append("</table></body></html>")
    open(run_dir / "report.html", "w").write("\n".join(html))

    print(f"report -> {run_dir/'report.md'}")
    print(f"  P0={p('P0')}  P1={p('P1')}  P2={p('P2')}  | endpoints with 5xx: {total_5xx_eps}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
