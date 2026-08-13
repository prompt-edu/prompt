"""api-stress-test — orchestrate one full stress + fuzz run against the ISOLATED
prompt-stress stack.

  preflight (all 7 servers + keycloak healthy)
  -> seed fixtures (2 courses, phases, student, per-service data, kc roles)
  -> mint tokens (AFTER role assignment)
  -> k6 scenarios (smoke, load, spike, soak, exhaustion) -> raw/k6_<scen>.json
  -> python fuzzer (inputs, authz, idor, files, slowloris)
  -> build prioritized report (report.md / report.html / findings.json)
  -> teardown the courses this run created (unless --keep-fixtures)

This is the Python port of the former run.sh orchestration; run.sh is now a thin
shim over this entry point, so `./run.sh …` and `uv run api-stress-test …` are
equivalent.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from datetime import datetime
from pathlib import Path

import httpx

CORE_URL = "http://localhost:18089"
KC_URL = "http://localhost:18081"

# Same set the old run.sh preflight polled: the 7 phase/core servers plus the
# isolated Keycloak realm's discovery document.
PREFLIGHT_URLS = [
    f"{CORE_URL}/api/hello",
    "http://localhost:18083/team-allocation/api/info",
    "http://localhost:18084/self-team-allocation/api/info",
    "http://localhost:18085/assessment/api/info",
    "http://localhost:18086/example-service/api/info",
    "http://localhost:18087/interview/api/info",
    "http://localhost:18088/certificate/api/info",
    f"{KC_URL}/realms/prompt/.well-known/openid-configuration",
]


def stress_dir() -> Path:
    """Locate the api-stress/ root (holds catalog/, lib/, k6/, …).

    Works whether the project is installed editable (``__file__`` under the
    source tree) or not (fall back to $STRESS_DIR / the current directory).
    """
    candidates = [
        Path(__file__).resolve().parent.parent,
        Path(os.environ["STRESS_DIR"]) if os.environ.get("STRESS_DIR") else None,
        Path.cwd(),
    ]
    for c in candidates:
        if c and (c / "catalog" / "merge_catalog.py").exists():
            return c
    return Path(__file__).resolve().parent.parent


STRESS_DIR = stress_dir()
PY = sys.executable


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(
        prog="api-stress-test",
        description="Run one full stress + fuzz pass against the isolated prompt-stress stack.",
    )
    p.add_argument(
        "--intensity",
        choices=["gentle", "medium", "brutal"],
        default="medium",
        help="k6 load level and fuzz body size (default: medium).",
    )
    p.add_argument(
        "--scenarios",
        default="smoke load spike soak exhaustion",
        help='space-separated subset of "smoke load spike soak exhaustion".',
    )
    p.add_argument(
        "--smoke-only", action="store_true", help="shorthand for --scenarios smoke."
    )
    p.add_argument(
        "--no-exhaustion", action="store_true", help="drop the exhaustion scenario."
    )
    p.add_argument(
        "--keep-fixtures",
        action="store_true",
        help="do not delete the courses this run created.",
    )
    return p.parse_args(argv)


def resolve_scenarios(args: argparse.Namespace) -> list[str]:
    if args.smoke_only:
        return ["smoke"]
    scenarios = args.scenarios.split()
    if args.no_exhaustion:
        scenarios = [s for s in scenarios if s != "exhaustion"]
    return scenarios


def require_k6() -> str:
    from shutil import which

    k6 = which("k6")
    if not k6:
        sys.exit("k6 not found on PATH (install: brew install k6)")
    return k6


def ensure_stress_env() -> None:
    # stress.env is gitignored; create it from the committed example on first run
    # so the suite works out of the box.
    env, example = STRESS_DIR / "stress.env", STRESS_DIR / "stress.env.example"
    if not env.exists() and example.exists():
        env.write_text(example.read_text())
        print("created stress.env from stress.env.example")


def run_step(argv: list[str], err: str) -> None:
    """Subprocess a leaf python script with the current (venv) interpreter."""
    if subprocess.run([PY, *argv], cwd=STRESS_DIR).returncode != 0:
        sys.exit(err)


def preflight() -> None:
    print("==> preflight")
    ok = True
    with httpx.Client(timeout=5) as client:
        for url in PREFLIGHT_URLS:
            try:
                code = client.get(url).status_code
            except httpx.HTTPError:
                code = 0
            if code == 200:
                print(f"  ok: {url}")
            else:
                print(f"  UNHEALTHY ({code}): {url}")
                ok = False
    if not ok:
        print("Stack not healthy. Bring it up with:")
        print(
            "  docker compose --env-file api-stress/stress.env -f docker-compose.yml "
            "-f api-stress/docker-compose.stress.yml -p prompt-stress up -d"
        )
        sys.exit(1)


def ensure_long_token_lifespan() -> None:
    # Access tokens must outlive a full run (default realm lifespan is 300s).
    # Idempotent; only touches the isolated realm.
    print("==> ensure long token lifespan")
    try:
        with httpx.Client(timeout=15) as client:
            token = client.post(
                f"{KC_URL}/realms/master/protocol/openid-connect/token",
                data={
                    "grant_type": "password",
                    "client_id": "admin-cli",
                    "username": "admin",
                    "password": "admin",
                },
            ).json()["access_token"]
            client.put(
                f"{KC_URL}/admin/realms/prompt",
                headers={
                    "Authorization": f"Bearer {token}",
                    "Content-Type": "application/json",
                },
                json={
                    "accessTokenLifespan": 3600,
                    "ssoSessionIdleTimeout": 7200,
                    "ssoSessionMaxLifespan": 7200,
                },
            )
        print("  accessTokenLifespan -> 3600s")
    except Exception as exc:  # noqa: BLE001 - best effort, never fatal
        print("  WARN could not bump token lifespan:", exc)


def raise_fd_limit() -> None:
    # Raise the fd limit so a localhost connection flood stresses the SERVER, not
    # the runner (ephemeral-port / fd exhaustion on the client side).
    try:
        import resource

        _, hard = resource.getrlimit(resource.RLIMIT_NOFILE)
        target = min(16384, hard) if hard != resource.RLIM_INFINITY else 16384
        resource.setrlimit(resource.RLIMIT_NOFILE, (target, hard))
    except Exception:  # noqa: BLE001 - not available on every platform
        pass


def catalog_count() -> int:
    try:
        data = json.loads((STRESS_DIR / "catalog" / "endpoints.json").read_text())
        return int(data.get("count", 0))
    except Exception:  # noqa: BLE001
        return 0


def run_k6(k6_bin: str, scenario: str, script: Path, run_dir: Path, intensity: str) -> None:
    print(f"==> k6: {scenario} (intensity={intensity})")
    env = {
        **os.environ,
        "STRESS_DIR": str(STRESS_DIR),
        "RUN_DIR": str(run_dir),
        "INTENSITY": intensity,
        "SCENARIO": scenario,
    }
    log = run_dir / "raw" / f"k6_{scenario}.log"
    with open(log, "w") as log_file:
        rc = subprocess.run(
            [
                k6_bin,
                "run",
                "--quiet",
                "--no-usage-report",
                "--out",
                f"json={run_dir}/raw/k6_{scenario}.json",
                str(script),
            ],
            env=env,
            stdout=log_file,
            stderr=subprocess.STDOUT,
        ).returncode
    print(f"    exit={rc} log={log}")


def teardown(run_dir: Path) -> None:
    print("==> teardown (delete this run's courses)")
    admin = json.loads((run_dir / "tokens.json").read_text())["tokens"].get("admin", "")
    fixtures = json.loads((run_dir / "fixtures.json").read_text())
    with httpx.Client(timeout=15) as client:
        for key in ("course_a", "course_b"):
            cid = fixtures.get(key, {}).get("id", "")
            if not cid:
                continue
            resp = client.delete(
                f"{CORE_URL}/api/courses/{cid}",
                headers={"Authorization": f"Bearer {admin}"},
            )
            print(f"  deleted course {cid} -> {resp.status_code}")


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    scenarios = resolve_scenarios(args)
    k6_bin = require_k6()
    ensure_stress_env()

    print("==> regenerate endpoint catalog from partials")
    run_step(["catalog/merge_catalog.py"], "catalog merge failed")

    preflight()
    ensure_long_token_lifespan()
    raise_fd_limit()

    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    run_dir = STRESS_DIR / "reports" / timestamp
    (run_dir / "raw").mkdir(parents=True, exist_ok=True)
    print(f"==> run dir: {run_dir}")

    meta = {
        "started": timestamp,
        "intensity": args.intensity,
        "scenarios": " ".join(scenarios),
        "stack": "prompt-stress (isolated, keycloak 26.4.7)",
        "catalog_count": catalog_count(),
        "core_url": CORE_URL,
    }
    (run_dir / "meta.json").write_text(json.dumps(meta, indent=2))

    print("==> seed fixtures")
    run_step(["lib/fixtures.py", str(run_dir)], "fixtures failed")
    print("==> mint tokens (post role-assignment)")
    run_step(["lib/auth.py", str(run_dir / "tokens.json")], "tokens failed")

    for scenario in scenarios:
        if scenario in ("smoke", "load", "spike", "soak"):
            run_k6(k6_bin, scenario, STRESS_DIR / "k6" / "scenario.js", run_dir, args.intensity)
        elif scenario == "exhaustion":
            run_k6(k6_bin, "exhaustion", STRESS_DIR / "k6" / "exhaustion.js", run_dir, args.intensity)
        else:
            print(f"  skip unknown scenario: {scenario}")

    print("==> re-mint tokens before fuzz (insurance against expiry during long k6 phase)")
    refreshed = subprocess.run(
        [PY, "lib/auth.py", str(run_dir / "tokens.json")],
        cwd=STRESS_DIR,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    if refreshed.returncode == 0:
        print("    tokens refreshed")

    print("==> fuzz")
    max_body_mb = {"brutal": 16, "gentle": 1}.get(args.intensity, 4)
    fuzz = subprocess.run(
        [PY, "fuzz/fuzz.py", str(run_dir), "--max-body-mb", str(max_body_mb)],
        cwd=STRESS_DIR,
    )
    if fuzz.returncode != 0:
        print("  fuzz returned nonzero (continuing)")

    print("==> build report")
    subprocess.run([PY, "report/build_report.py", str(run_dir)], cwd=STRESS_DIR)

    if not args.keep_fixtures:
        teardown(run_dir)

    print("")
    print("==================== DONE ====================")
    print(f"Report:   {run_dir}/report.md")
    print(f"Findings: {run_dir}/findings.json")
    print(f"HTML:     {run_dir}/report.html")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
