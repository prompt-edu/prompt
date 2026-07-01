#!/usr/bin/env python3
"""Security + robustness fuzzer for the PROMPT API.

Consumes the SAME catalog the k6 load uses, plus the run's fixtures/tokens, and
probes every endpoint along four axes, emitting structured findings the report
turns into actionable bugs:

  1. inputs  - malformed path UUIDs + malformed/oversized/nested JSON bodies.
               A 5xx where a 4xx is expected = robustness bug.
  2. authz   - no token / forged-signature token / wrong-role token against
               protected endpoints. A 2xx/3xx = auth bypass.
  3. idor    - course-A-scoped token reaching course-B resources. A 2xx with a
               body = cross-tenant data exposure.
  4. files   - presign/complete abuse (oversized, wrong mime, orphan/complete
               of never-uploaded key).

Plus a small raw-socket slowloris probe (held partial requests) to see whether a
server ties up workers - the one DoS lane k6 cannot express.

Every finding carries a ready-to-paste curl repro. Writes findings.json + a
raw per-request log to the run dir.

Usage: fuzz.py <run_dir> [--max-body-mb N]
"""
from __future__ import annotations

import base64
import json
import pathlib
import socket
import sys
import time
import uuid as uuidlib

import httpx

HERE = pathlib.Path(__file__).parent
LIB = HERE.parent / "lib"
sys.path.insert(0, str(LIB))
import auth  # noqa: E402

SERVICES = json.load(open(LIB / "services.json"))
CATALOG = json.load(open(HERE.parent / "catalog" / "endpoints.json"))["endpoints"]
FALLBACK_UUID = "00000000-0000-0000-0000-000000000000"

SERVICE_PHASE_TYPE = {
    "interview": "interview",
    "team_allocation": "team allocation",
    "self_team_allocation": "self team allocation",
    "assessment": "assessment",
    "certificate": "certificate",
    "template_server": "template",
    "core": "application",
}
ROLE_TO_USER = {
    "PromptAdmin": "admin",
    "PromptLecturer": "lecturer",
    "CourseLecturer": "course-lecturer",
    "CourseEditor": "course-editor",
    "CourseStudent": "student",
}
# what role each seeded user effectively carries (for wrong-role selection)
USER_ROLES = {
    "admin": {"PromptAdmin", "PromptLecturer"},
    "lecturer": {"PromptLecturer"},
    "course-lecturer": {"CourseLecturer"},
    "course-editor": {"CourseEditor"},
    "student": {"CourseStudent"},
}


class Fuzzer:
    def __init__(self, run_dir: pathlib.Path, max_body_mb: int = 4):
        self.run_dir = run_dir
        self.max_body_mb = max_body_mb
        self.fx = json.load(open(run_dir / "fixtures.json"))
        self.tokens = json.load(open(run_dir / "tokens.json")).get("tokens", {})
        self.findings: list[dict] = []
        self.raw: list[dict] = []
        self.client = httpx.Client(timeout=30.0)

    # ---------- helpers ----------
    def base(self, service):
        return SERVICES["services"][service]["base_url"]

    def phase_for(self, service, course="a"):
        want = SERVICE_PHASE_TYPE.get(service, "application")
        phases = self.fx["phases_a"] if course == "a" else self.fx["phases_b"]
        if phases.get(want, {}).get("id"):
            return phases[want]["id"]
        if phases.get("application", {}).get("id"):
            return phases["application"]["id"]
        for p in phases.values():
            if p.get("id"):
                return p["id"]
        return FALLBACK_UUID

    def resolve(self, ep, course="a", override=None):
        override = override or {}
        ca = self.fx["course_a"]["id"] if course == "a" else self.fx["course_b"]["id"]
        student = (self.fx.get("student") or {}).get("student_id") or FALLBACK_UUID
        part = (self.fx.get("student") or {}).get("participation_id") or FALLBACK_UUID
        path = ep["path"]
        for p in ep.get("pathParams", []):
            if p in override:
                val = override[p]
            elif p == "coursePhaseID":
                val = self.phase_for(ep["service"], course)
            elif p == "courseID":
                val = ca
            elif p in ("course_participation_id", "courseParticipationID"):
                val = part
            elif p in ("student-uuid", "studentID", "studentId"):
                val = student
            elif p == "searchString":
                val = "a"
            elif p == "groupName":
                val = "Prompt"
            elif p == "uuid":
                if "/courses/" in path:
                    val = ca
                elif "/course_phases/" in path:
                    val = self.phase_for("core", course)
                elif "/students/" in path:
                    val = student
                else:
                    val = ca
            else:
                val = FALLBACK_UUID
            path = path.replace("{" + p + "}", str(val))
        return self.base(ep["service"]) + path

    def hdr(self, token=None):
        h = {"Content-Type": "application/json"}
        if token:
            h["Authorization"] = f"Bearer {token}"
        return h

    def curl(self, method, url, token=None, body=None):
        parts = [f"curl -sS -X {method}", f"'{url}'"]
        if token:
            parts.append(f"-H 'Authorization: Bearer <{token}_TOKEN>'")
        if body is not None:
            parts.append("-H 'Content-Type: application/json'")
            b = body if isinstance(body, str) else json.dumps(body)
            if len(b) > 200:
                b = b[:200] + "...<truncated>"
            parts.append(f"--data '{b}'")
        return " ".join(parts)

    def do(self, method, url, token=None, body=None, label=""):
        try:
            r = self.client.request(method, url, headers=self.hdr(token), content=(
                body if isinstance(body, (str, bytes)) else (json.dumps(body) if body is not None else None)
            ))
            rec = {"label": label, "method": method, "url": url, "status": r.status_code,
                   "ms": r.elapsed.total_seconds() * 1000, "len": len(r.content)}
            self.raw.append(rec)
            return r.status_code, r.text[:400], rec
        except Exception as e:  # noqa: BLE001
            rec = {"label": label, "method": method, "url": url, "status": 0, "error": str(e)[:200]}
            self.raw.append(rec)
            return 0, str(e)[:200], rec

    def finding(self, **kw):
        kw.setdefault("severity", "medium")
        self.findings.append(kw)

    @staticmethod
    def is_empty_body(txt):
        # A cross-course 200 only proves data exposure if the body actually
        # carries data. Normalize + parse instead of string-matching the raw
        # (truncated) text, so wrapped-empty bodies ({"data":[]}), constant
        # blobs of only-empty collections, or whitespace-padded [] / {} / null
        # don't get reported as a spurious P0 IDOR.
        if txt is None:
            return True
        s = txt.strip()
        if not s:
            return True
        try:
            v = json.loads(s)
        except ValueError:
            return False  # unparseable (e.g. truncated real payload) -> treat as data
        return Fuzzer._empty_value(v)

    @staticmethod
    def _empty_value(v):
        if v is None or v == [] or v == {} or v == "":
            return True
        if isinstance(v, dict):
            payload_keys = {"data", "items", "results", "entries"}
            present = payload_keys.intersection(v)
            if present and all(Fuzzer._empty_value(v[k]) for k in present):
                return True  # empty paginated envelope: metadata like total/success is not content
            return all(Fuzzer._empty_value(x) for x in v.values())
        return False

    def token_user(self, name):
        return self.tokens.get(name)

    def forged_token(self):
        # take a real token, corrupt the signature -> must be rejected (401)
        t = self.tokens.get("admin") or next(iter(self.tokens.values()), None)
        if not t:
            return None
        head, payload, _sig = t.split(".")
        return f"{head}.{payload}.AAAAinvalidsignatureAAAA"

    def none_aud_token(self):
        # alg=none unsigned token forgery attempt. The claims must match what a
        # genuinely-accepted token needs so the probe can actually flip to 2xx on
        # a vulnerable server: aud=prompt-server (checkAudience, middleware.go:142)
        # and the real admin role string PROMPT_Admin under resource_access
        # (checkKeycloakRoles, middleware.go:175). Omitting aud makes every role
        # get dropped, masking a real alg=none bypass as a false negative.
        header = base64.urlsafe_b64encode(b'{"alg":"none","typ":"JWT"}').decode().rstrip("=")
        payload = base64.urlsafe_b64encode(
            json.dumps({"sub": "attacker", "azp": "prompt-client", "aud": "prompt-server",
                        "resource_access": {"prompt-server": {"roles": ["PROMPT_Admin"]}}}).encode()
        ).decode().rstrip("=")
        return f"{header}.{payload}."

    # ---------- axis 1: input fuzzing ----------
    def bad_uuid_values(self):
        return [
            "not-a-uuid",
            "11111111-1111-1111-1111-11111111111Z",
            "%2e%2e%2f%2e%2e%2f",
            "' OR '1'='1",
            "<script>alert(1)</script>",
            "A" * 5000,
            "00000000-0000-0000-0000-000000000000\x00",
        ]

    def bad_bodies(self):
        big = "A" * (self.max_body_mb * 1024 * 1024)
        nested = "1"
        for _ in range(6000):
            nested = '{"a":' + nested + "}"
        return [
            ("empty-object", "{}"),
            ("null", "null"),
            ("array", "[]"),
            ("not-json", "{not json,,,"),
            ("wrong-types", json.dumps({"name": 12345, "startDate": True, "ects": "lots"})),
            ("sqli", json.dumps({"name": "'; DROP TABLE courses;--"})),
            ("xss", json.dumps({"name": "<img src=x onerror=alert(1)>"})),
            ("huge-string", json.dumps({"name": big})),
            ("deeply-nested", nested),
        ]

    def fuzz_inputs(self):
        admin = self.token_user("admin")
        # 1a: malformed path UUID on a representative set of GET-by-id endpoints
        for ep in CATALOG:
            params = ep.get("pathParams", [])
            idparam = next((p for p in params if p in (
                "uuid", "coursePhaseID", "courseParticipationID", "course_participation_id",
            )), None)
            if not idparam or ep["method"] != "GET":
                continue
            for bad in self.bad_uuid_values()[:4]:
                url = self.resolve(ep, override={idparam: bad})
                st, txt, _ = self.do(ep["method"], url, token=admin, label="bad-uuid")
                if st >= 500 or st == 0:
                    self.finding(axis="inputs", endpoint=ep["id"], method=ep["method"],
                                 url=url, observed=f"HTTP {st}", expected="HTTP 400 (bad id)",
                                 detail=f"malformed path param {idparam!r} -> server error",
                                 evidence=txt, repro=self.curl(ep["method"], url, "admin"),
                                 severity="high")
                    break
        # 1b: malformed bodies on write endpoints (real ids so they reach handler)
        for ep in CATALOG:
            if ep["method"] not in ("POST", "PUT"):
                continue
            url = self.resolve(ep)
            for cls, body in self.bad_bodies():
                st, txt, _ = self.do(ep["method"], url, token=admin, body=body, label=f"body:{cls}")
                if st >= 500 or st == 0:
                    self.finding(axis="inputs", endpoint=ep["id"], method=ep["method"], url=url,
                                 observed=f"HTTP {st}", expected="HTTP 400 (bad body)",
                                 detail=f"{cls} body -> server error", evidence=txt,
                                 repro=self.curl(ep["method"], url, "admin", body),
                                 severity="high" if cls in ("huge-string", "deeply-nested") else "medium")

    # ---------- axis 2: authz ----------
    def fuzz_auth(self):
        forged = self.forged_token()
        none_tok = self.none_aud_token()
        for ep in CATALOG:
            roles = ep.get("roles", [])
            if not roles or roles[0] in ("public",):
                continue
            url = self.resolve(ep)
            method = ep["method"]
            # destructive/writes: only probe with GET-safe semantics? we still send
            # them but to course A real ids; acceptable - they are auth-rejected
            # before mutating in the bypass case (the point is they SHOULD reject).
            # no token
            st, txt, _ = self.do(method, url, token=None, label="no-token")
            if 200 <= st < 400:
                self.finding(axis="authz", endpoint=ep["id"], method=method, url=url,
                             observed=f"HTTP {st} with NO token", expected="401",
                             detail="protected endpoint reachable without authentication",
                             evidence=txt, repro=self.curl(method, url), severity="critical")
            # forged signature
            st, txt, _ = self.do(method, url, token=forged, label="forged-sig")
            if 200 <= st < 400:
                self.finding(axis="authz", endpoint=ep["id"], method=method, url=url,
                             observed=f"HTTP {st} with forged-signature token", expected="401",
                             detail="token signature not validated", evidence=txt,
                             repro=self.curl(method, url, "FORGED"), severity="critical")
            # alg=none unsigned
            st, txt, _ = self.do(method, url, token=none_tok, label="alg-none")
            if 200 <= st < 400:
                self.finding(axis="authz", endpoint=ep["id"], method=method, url=url,
                             observed=f"HTTP {st} with alg=none unsigned token", expected="401",
                             detail="accepts unsigned (alg=none) JWT", evidence=txt,
                             repro=self.curl(method, url, "ALGNONE"), severity="critical")
            # privilege escalation: the LOWEST-privilege actor (a course student)
            # hitting an endpoint that does NOT permit CourseStudent must be denied.
            # We deliberately do NOT probe with admin/lecturer as "wrong role" -
            # PromptAdmin/PromptLecturer are global privileged roles legitimately
            # allowed broadly, and "my-*" endpoints return caller-scoped data, so
            # those produce false positives rather than real bypasses.
            student_excluded = "CourseStudent" not in roles and roles[0] != "custom"
            is_self_scoped = url.rstrip("/").endswith(("/self", "my-results", "my-evaluations",
                                                       "my-completions", "my-feedback",
                                                       "my-action-items", "my-assignment",
                                                       "my-grade-suggestion"))
            if student_excluded and not is_self_scoped and self.tokens.get("student"):
                st, txt, _ = self.do(method, url, token=self.tokens["student"], label="escalation:student")
                if 200 <= st < 400:
                    self.finding(axis="authz", endpoint=ep["id"], method=method, url=url,
                                 observed=f"HTTP {st} with course-STUDENT token",
                                 expected="403 (endpoint requires " + ",".join(roles) + ", not CourseStudent)",
                                 detail="privilege escalation: course student reached a higher-privilege endpoint",
                                 evidence=txt, repro=self.curl(method, url, "student"), severity="high")

    # ---------- axis 3: IDOR (cross-course) ----------
    def fuzz_idor(self):
        # Both actors are scoped to course A only. Reaching course B = horizontal
        # access-control gap. We test the student (CourseStudent endpoints) and the
        # course-lecturer (management endpoints) so we catch leaks at both levels.
        actors = [
            ("student", self.token_user("student"), {"CourseStudent"}),
            ("course-lecturer", self.token_user("course-lecturer"), {"CourseLecturer", "CourseEditor"}),
        ]
        for ep in CATALOG:
            if ep["method"] != "GET":
                continue
            if "coursePhaseID" not in ep.get("pathParams", []) and "courseID" not in ep.get("pathParams", []):
                continue
            roles = set(ep.get("roles", []))
            if not roles or "public" in roles or "custom" in roles:
                continue
            if ep["path"].endswith("/hello") or ep["path"].endswith("/info"):
                continue
            url_b = self.resolve(ep, course="b")
            for name, tok, actor_roles in actors:
                if not tok or roles.isdisjoint(actor_roles):
                    continue  # this actor wouldn't be allowed even in its own course
                st, txt, _ = self.do(ep["method"], url_b, token=tok, label=f"idor-{name}-courseB")
                if st == 200 and not self.is_empty_body(txt):
                    self.finding(axis="idor", endpoint=ep["id"], method=ep["method"], url=url_b,
                                 observed=f"HTTP 200 with body, {name} (course-A only) reading course B",
                                 expected="403/404 (actor not enrolled/assigned in course B)",
                                 detail="cross-course broken access control (BOLA/IDOR): per-course "
                                        "membership not enforced; core denies the same access (403)",
                                 evidence=txt[:300], repro=self.curl(ep["method"], url_b, name),
                                 severity="high")
                    break  # one actor proving it is enough per endpoint

    # ---------- axis 4: file abuse (best-effort) ----------
    def fuzz_files(self):
        core = self.base("core")
        appid = self.phase_for("core")
        admin = self.token_user("admin")
        # complete an upload that never happened
        url = f"{core}/api/apply/{appid}/files/complete"
        st, txt, _ = self.do("POST", url, token=None, body={"fileId": str(uuidlib.uuid4()), "key": "../../etc/passwd"},
                             label="file-complete-orphan")
        if st >= 500:
            self.finding(axis="files", endpoint="core.POST.apply-files-complete", method="POST", url=url,
                         observed=f"HTTP {st}", expected="4xx", detail="completing a never-uploaded key -> server error",
                         evidence=txt, repro=self.curl("POST", url, None, {"key": "../../etc/passwd"}),
                         severity="medium")
        # presign with path traversal filename
        url = f"{core}/api/apply/{appid}/files/presign"
        st, txt, _ = self.do("POST", url, token=None, body={"filename": "../../../../etc/passwd", "contentType": "text/plain", "size": 10},
                             label="file-presign-traversal")
        if st >= 500:
            self.finding(axis="files", endpoint="core.POST.apply-files-presign", method="POST", url=url,
                         observed=f"HTTP {st}", expected="4xx", detail="path-traversal filename on presign -> server error",
                         evidence=txt, repro=self.curl("POST", url, None, {"filename": "../../../../etc/passwd"}),
                         severity="medium")

    # ---------- axis 5: slowloris (raw socket, one server) ----------
    def slowloris(self, host="localhost", port=18089, holders=30, hold_s=6):
        """Open many connections, send headers slowly, never finish. Measures
        whether normal requests still get served while connections are held."""
        socks = []
        try:
            for _ in range(holders):
                try:
                    s = socket.create_connection((host, port), timeout=3)
                    s.sendall(b"POST /api/courses/ HTTP/1.1\r\nHost: localhost\r\nContent-Length: 1000000\r\n")
                    socks.append(s)
                except Exception:  # noqa: BLE001
                    break
            # while holding, can we still get a fast health response?
            t0 = time.time()
            try:
                r = httpx.get(f"http://{host}:{port}/api/hello", timeout=5)
                healthy_ms = (time.time() - t0) * 1000
                ok = r.status_code == 200
            except Exception as e:  # noqa: BLE001
                healthy_ms = (time.time() - t0) * 1000
                ok = False
                _ = e
            time.sleep(hold_s)
            if not ok or healthy_ms > 3000:
                self.finding(axis="dos", endpoint="core.GET.hello", method="GET",
                             url=f"http://{host}:{port}/api/hello",
                             observed=f"health {'unreachable' if not ok else f'slow {healthy_ms:.0f}ms'} while {len(socks)} slowloris conns held",
                             expected="health stays fast under held/partial connections",
                             detail="server may not bound slow-client connections (slowloris exposure)",
                             evidence=f"{len(socks)} partial connections held for {hold_s}s",
                             repro="see fuzz.py slowloris() - raw sockets sending partial headers",
                             severity="medium")
            else:
                self.raw.append({"label": "slowloris", "note": f"health stayed fast ({healthy_ms:.0f}ms) with {len(socks)} held conns"})
        finally:
            for s in socks:
                try:
                    s.close()
                except Exception:  # noqa: BLE001
                    pass

    def run(self):
        print("  [1/5] input fuzzing (path uuids + bodies)...")
        self.fuzz_inputs()
        print(f"      findings so far: {len(self.findings)}")
        print("  [2/5] auth matrix (no/forged/alg-none/wrong-role)...")
        self.fuzz_auth()
        print(f"      findings so far: {len(self.findings)}")
        print("  [3/5] IDOR (cross-course)...")
        self.fuzz_idor()
        print("  [4/5] file abuse...")
        self.fuzz_files()
        print("  [5/5] slowloris...")
        self.slowloris()
        return self.findings


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: fuzz.py <run_dir> [--max-body-mb N]", file=sys.stderr)
        return 2
    run_dir = pathlib.Path(sys.argv[1])
    mb = 4
    if "--max-body-mb" in sys.argv:
        mb = int(sys.argv[sys.argv.index("--max-body-mb") + 1])
    fz = Fuzzer(run_dir, max_body_mb=mb)
    findings = fz.run()
    json.dump(findings, open(run_dir / "fuzz_findings.json", "w"), indent=2)
    json.dump(fz.raw, open(run_dir / "fuzz_raw.json", "w"), indent=2)
    by_axis = {}
    for f in findings:
        by_axis[f["axis"]] = by_axis.get(f["axis"], 0) + 1
    print(f"\nwrote {len(findings)} findings -> {run_dir/'fuzz_findings.json'}")
    for a, n in sorted(by_axis.items()):
        print(f"  {a:8s} {n}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
