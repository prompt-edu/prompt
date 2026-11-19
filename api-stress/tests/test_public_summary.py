import json

import public_summary

P0 = {
    "priority": "P0",
    "kind": "idor",
    "endpoint": "assessment.get.secret",
    "method": "GET",
    "title": "cross-course read",
    "observed": "200 with course B data",
}
P1 = {
    "priority": "P1",
    "kind": "server-error",
    "endpoint": "core.get.course",
    "method": "GET",
    "title": "5xx on a valid read request",
    "observed": "statuses {'500': 1}",
}
P2 = {
    "priority": "P2",
    "kind": "inputs",
    "endpoint": "core.put.course",
    "method": "PUT",
    "title": "5xx on malformed | piped input",
    "observed": "x" * 500,
}
META = {"intensity": "gentle", "scenarios": "smoke", "catalog_count": 252}


def test_redact_drops_p0_and_counts_it():
    public, p0 = public_summary.redact([P0, P1, P2, P0])
    assert public == [P1, P2]
    assert p0 == 2


def test_render_never_leaks_p0_details():
    summary = public_summary.render([P0, P1, P2], META)
    for value in P0.values():
        if value not in ("P0", "GET"):
            assert value not in summary
    assert "| P0 (auth bypass / IDOR / outage) | 1 |" in summary
    assert "[!CAUTION]" in summary


def test_render_lists_p1_and_p2_in_full():
    summary = public_summary.render([P1, P2], META)
    assert "`core.get.course`" in summary
    assert "<summary>P1 - High (1)</summary>" in summary
    assert "<summary>P2 - Robustness / Performance (1)</summary>" in summary
    # Pipes are escaped and long observations are cut, so the table stays intact.
    assert "malformed \\| piped" in summary
    assert "x" * public_summary.OBSERVED_MAX in summary
    assert "x" * (public_summary.OBSERVED_MAX + 1) not in summary
    assert "[!CAUTION]" not in summary


def test_main_writes_redacted_files(tmp_path, monkeypatch):
    run_dir, out_dir = tmp_path / "run", tmp_path / "public"
    run_dir.mkdir()
    (run_dir / "findings.json").write_text(json.dumps([P0, P1]))
    (run_dir / "meta.json").write_text(json.dumps(META))
    monkeypatch.setattr("sys.argv", ["public_summary.py", str(run_dir), str(out_dir)])

    assert public_summary.main() == 0
    assert json.loads((out_dir / "findings.json").read_text()) == [P1]
    assert P0["endpoint"] not in (out_dir / "summary.md").read_text()
