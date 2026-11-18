import json

import pytest

from stress import cli

FINDINGS = [{"priority": "P1"}, {"priority": "P2"}, {"priority": "P2"}]


@pytest.mark.parametrize(("threshold", "expected"), [("P0", 0), ("P1", 1), ("P2", 3)])
def test_count_at_or_above_includes_more_severe_priorities(threshold, expected):
    assert cli.count_at_or_above(FINDINGS, threshold) == expected


def write_run(run_dir, findings):
    (run_dir / "fuzz_findings.json").write_text("[]")
    (run_dir / "findings.json").write_text(json.dumps(findings))


def test_gate_passes_without_findings_at_threshold(tmp_path):
    write_run(tmp_path, FINDINGS)
    assert cli.gate(tmp_path, "P0") == 0


def test_gate_fails_on_finding_at_threshold(tmp_path):
    write_run(tmp_path, [*FINDINGS, {"priority": "P0"}])
    assert cli.gate(tmp_path, "P0") == 1


def test_gate_fails_when_the_report_is_missing(tmp_path):
    write_run(tmp_path, FINDINGS)
    (tmp_path / "findings.json").unlink()
    assert cli.gate(tmp_path, "P2") == 1


def test_gate_fails_when_the_fuzz_findings_are_missing(tmp_path):
    # build_report.py still writes findings.json after a fuzzer crash.
    write_run(tmp_path, FINDINGS)
    (tmp_path / "fuzz_findings.json").unlink()
    assert cli.gate(tmp_path, "P2") == 1


def test_gate_fails_when_the_fuzzer_exited_nonzero(tmp_path):
    write_run(tmp_path, FINDINGS)
    assert cli.gate(tmp_path, "P2", fuzz_returncode=1) == 1


def test_parse_args_gate_flags_default_off():
    args = cli.parse_args([])
    assert args.fail_on is None
    assert args.public_summary is None
