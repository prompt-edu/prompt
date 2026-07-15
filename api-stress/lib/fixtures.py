#!/usr/bin/env python3
"""Bootstrap deterministic, ISOLATED test data via the live API.

Docker DBs start empty (migrations only). This seeds, using a PromptAdmin token:
  - two courses A and B (B exists so the auth fuzzer can probe cross-course IDOR)
  - one course phase per available phase type in each course
  - a student matching the Keycloak `student` user, enrolled in course A
  - best-effort per-microservice data so read endpoints return non-empty
  - best-effort Keycloak group assignment for course-lecturer / course-editor

Writes fixtures.json (all generated UUIDs + which seeding succeeded) into the run
dir. Designed to be RE-RUNNABLE: every course is uniquely named per run, and all
non-essential steps are best-effort (failures are recorded, not fatal) so the
suite always gets a usable fixture set.

Usage: fixtures.py <run_dir>
"""
from __future__ import annotations

import json
import pathlib
import sys
import time

import httpx

HERE = pathlib.Path(__file__).parent
sys.path.insert(0, str(HERE))
import auth  # noqa: E402

SERVICES = json.load(open(HERE / "services.json"))
CORE = SERVICES["services"]["core"]["base_url"]
KC = SERVICES["keycloak"]

# Keycloak `student` user identity (keycloakConfig.json) — must match for the
# core role resolver to grant CourseStudent on the enrolled course.
STUDENT_IDENTITY = {
    "firstName": "Stress",
    "lastName": "Student",
    "email": "pgdp_enjoyer@example.com",
    "matriculationNumber": "00000005",
    "universityLogin": "no42tum",
    "hasUniversityAccount": True,
    "gender": "prefer_not_to_say",
    "nationality": "DE",
    "studyDegree": "master",
    "studyProgram": "Informatics",
    "currentSemester": 3,
}

# phase-type name (lower) -> which microservice it drives (for seeding + targeting)
TYPE_TO_SERVICE = {
    "application": "core",
    "interview": "interview",
    "team allocation": "team_allocation",
    "self team allocation": "self_team_allocation",
    "assessment": "assessment",
    "example": "example_server",
    "matching": "core",
}


class Boot:
    def __init__(self, run_dir: pathlib.Path):
        self.run_dir = run_dir
        self.log: list[str] = []
        self.fixtures: dict = {"seed_log": self.log, "seeded": {}}
        # stamp uniqueness from the run dir name (no Date.now allowed elsewhere,
        # but here we just need per-run uniqueness; run_dir name is a timestamp).
        self.tag = "stress-" + run_dir.name.replace("_", "")[-10:]
        admin = auth.fetch_token("admin", "admin")["access_token"]
        self.admin_h = {"Authorization": f"Bearer {admin}"}
        self.client = httpx.Client(timeout=20.0)

    def _log(self, msg: str):
        self.log.append(msg)
        print("  " + msg)

    def req(self, method: str, url: str, headers=None, json_body=None, ok=(200, 201)):
        r = self.client.request(method, url, headers=headers or self.admin_h, json=json_body)
        if r.status_code not in ok:
            raise RuntimeError(f"{method} {url} -> {r.status_code}: {r.text[:300]}")
        if r.text:
            try:
                return r.json()
            except Exception:  # noqa: BLE001
                return r.text
        return None

    # ---------- course + phases ----------
    def create_course(self, label: str) -> dict:
        body = {
            "name": f"STRESS-{label}",
            "semesterTag": f"{self.tag}-{label}".lower(),
            "courseType": "practical course",
            "ects": 5,
            "startDate": "2026-01-01",
            "endDate": "2026-06-30",
            "template": False,
            "restrictedData": {},
            "studentReadableData": {},
        }
        c = self.req("POST", f"{CORE}/api/courses/", json_body=body)
        cid = c["id"] if isinstance(c, dict) else None
        self._log(f"course {label} created: {cid} (semesterTag {body['semesterTag']})")
        return {"id": cid, "semesterTag": body["semesterTag"], "name": body["name"]}

    def phase_types(self) -> list[dict]:
        types = self.req("GET", f"{CORE}/api/course_phase_types")
        return types if isinstance(types, list) else []

    def create_phases(self, course_id: str, types: list[dict]) -> dict:
        phases = {}
        first = True
        for t in types:
            name = (t.get("name") or "").strip()
            tid = t.get("id")
            if not tid:
                continue
            body = {
                "courseID": course_id,
                "name": f"{name} Phase",
                "isInitialPhase": first,
                "coursePhaseTypeID": tid,
                "restrictedData": {},
                "studentReadableData": {},
            }
            try:
                p = self.req(
                    "POST", f"{CORE}/api/course_phases/course/{course_id}", json_body=body
                )
                pid = p["id"] if isinstance(p, dict) else None
                phases[name.lower()] = {"id": pid, "typeID": tid, "typeName": name}
                first = False
            except Exception as e:  # noqa: BLE001
                self._log(f"  phase '{name}' FAILED: {e}")
        self._log(f"created {len(phases)} phases for course {course_id}")
        return phases

    # ---------- student enrollment ----------
    def create_and_enroll_student(self, course_id: str) -> dict:
        out = {"student_id": None, "participation_id": None}
        try:
            s = self.req("POST", f"{CORE}/api/students/", json_body=STUDENT_IDENTITY)
            sid = s.get("id") if isinstance(s, dict) else None
            out["student_id"] = sid
            self._log(f"student created: {sid}")
        except Exception as e:  # noqa: BLE001
            self._log(f"student create FAILED (may already exist): {e}")
            # try to find existing by search
            try:
                found = self.req(
                    "GET", f"{CORE}/api/students/search/{STUDENT_IDENTITY['universityLogin']}"
                )
                if isinstance(found, list) and found:
                    out["student_id"] = found[0].get("id")
                    self._log(f"  found existing student: {out['student_id']}")
            except Exception:  # noqa: BLE001
                pass
        if out["student_id"]:
            try:
                p = self.req(
                    "POST",
                    f"{CORE}/api/courses/{course_id}/participations/enroll",
                    json_body={"courseID": course_id, "studentID": out["student_id"]},
                )
                out["participation_id"] = p.get("id") if isinstance(p, dict) else None
                self._log(f"enrolled student -> participation {out['participation_id']}")
            except Exception as e:  # noqa: BLE001
                self._log(f"enroll FAILED: {e}")
        return out

    # ---------- per-service best-effort seeding ----------
    def seed_services(self, phases: dict):
        seeded = self.fixtures["seeded"]
        S = SERVICES["services"]

        def phase_for(svc_type_name):
            return (phases.get(svc_type_name) or {}).get("id")

        # team allocation: team + skill
        ta = phase_for("team allocation")
        if ta:
            base = S["team_allocation"]["base_url"] + f"/team-allocation/api/course_phase/{ta}"
            seeded["team_allocation"] = {}
            for label, url, body in [
                ("team", f"{base}/team", [{"name": "Stress Team A"}]),
                ("skill", f"{base}/skill", [{"name": "Go"}]),
            ]:
                try:
                    self.req("POST", url, json_body=body)
                    seeded["team_allocation"][label] = "ok"
                except Exception as e:  # noqa: BLE001
                    seeded["team_allocation"][label] = f"fail: {str(e)[:120]}"

        # interview: slot
        iv = phase_for("interview")
        if iv:
            base = S["interview"]["base_url"] + f"/interview/api/course_phase/{iv}"
            seeded["interview"] = {}
            try:
                self.req(
                    "POST",
                    f"{base}/interview-slots",
                    json_body={
                        "startTime": "2026-02-01T09:00:00Z",
                        "endTime": "2026-02-01T09:30:00Z",
                        "duration": 30,
                        "location": "Room 1",
                    },
                )
                seeded["interview"]["slot"] = "ok"
            except Exception as e:  # noqa: BLE001
                seeded["interview"]["slot"] = f"fail: {str(e)[:120]}"

        # assessment: category + competency (schema shapes vary; best-effort)
        asm = phase_for("assessment")
        if asm:
            base = S["assessment"]["base_url"] + f"/assessment/api/course_phase/{asm}"
            seeded["assessment"] = {}
            try:
                self.req(
                    "POST",
                    f"{base}/category",
                    json_body={"name": "Stress Category", "shortName": "SC", "weight": 1},
                )
                seeded["assessment"]["category"] = "ok"
            except Exception as e:  # noqa: BLE001
                seeded["assessment"]["category"] = f"fail: {str(e)[:120]}"

    # ---------- keycloak group role assignment (best-effort) ----------
    def assign_course_roles(self, course):
        try:
            tok = httpx.post(
                f"{KC['base_url']}/realms/master/protocol/openid-connect/token",
                data={
                    "grant_type": "password",
                    "client_id": "admin-cli",
                    "username": KC["admin_user"],
                    "password": KC["admin_pass"],
                },
                timeout=15,
            )
            tok.raise_for_status()
            kch = {"Authorization": f"Bearer {tok.json()['access_token']}"}
            admin_base = f"{KC['base_url']}/admin/realms/{KC['realm']}"

            def user_id(username):
                u = httpx.get(
                    f"{admin_base}/users", params={"username": username, "exact": "true"},
                    headers=kch, timeout=15,
                ).json()
                return u[0]["id"] if u else None

            # find course subgroups Lecturer/Editor under /Prompt/<semester>-<name>
            groups = httpx.get(
                f"{admin_base}/groups", params={"search": "Prompt"}, headers=kch, timeout=15
            ).json()
            course_group_name = f"{course['semesterTag']}-{course['name']}"

            def find_subgroup(node, target_course, role):
                # walk group tree to find /Prompt/<course_group>/<role>
                for g in node:
                    sub = httpx.get(
                        f"{admin_base}/groups/{g['id']}/children", headers=kch, timeout=15
                    ).json() if g.get("subGroupCount", 1) else g.get("subGroups", [])
                    for cg in sub:
                        if cg["name"] == target_course:
                            roles = httpx.get(
                                f"{admin_base}/groups/{cg['id']}/children",
                                headers=kch, timeout=15,
                            ).json()
                            for rg in roles:
                                if rg["name"] == role:
                                    return rg["id"]
                return None

            out = {}
            for username, role in [("course-lecturer", "Lecturer"), ("course-editor", "Editor")]:
                uid = user_id(username)
                gid = find_subgroup(groups, course_group_name, role)
                if uid and gid:
                    r = httpx.put(
                        f"{admin_base}/users/{uid}/groups/{gid}", headers=kch, timeout=15
                    )
                    out[username] = "ok" if r.status_code in (204, 201) else f"http {r.status_code}"
                else:
                    out[username] = f"not-found (uid={bool(uid)}, gid={bool(gid)})"
            self._log(f"kc role assignment: {out}")
            return out
        except Exception as e:  # noqa: BLE001
            self._log(f"kc role assignment FAILED: {e}")
            return {"error": str(e)[:200]}

    def run(self) -> dict:
        types = self.phase_types()
        self._log(f"discovered {len(types)} phase types: {[t.get('name') for t in types]}")
        course_a = self.create_course("A")
        course_b = self.create_course("B")
        phases_a = self.create_phases(course_a["id"], types)
        phases_b = self.create_phases(course_b["id"], types)
        student = self.create_and_enroll_student(course_a["id"])
        self.seed_services(phases_a)
        kc = self.assign_course_roles(course_a)

        self.fixtures.update(
            {
                "tag": self.tag,
                "course_a": course_a,
                "course_b": course_b,
                "phases_a": phases_a,
                "phases_b": phases_b,
                "student": student,
                "kc_role_assignment": kc,
                "type_to_service": TYPE_TO_SERVICE,
            }
        )
        return self.fixtures


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: fixtures.py <run_dir>", file=sys.stderr)
        return 2
    run_dir = pathlib.Path(sys.argv[1])
    run_dir.mkdir(parents=True, exist_ok=True)
    boot = Boot(run_dir)
    fx = boot.run()
    dest = run_dir / "fixtures.json"
    json.dump(fx, open(dest, "w"), indent=2)
    print(f"\nwrote fixtures -> {dest}")
    print(f"  course A: {fx['course_a']['id']}  ({len(fx['phases_a'])} phases)")
    print(f"  course B: {fx['course_b']['id']}  ({len(fx['phases_b'])} phases)")
    print(f"  student:  {fx['student']['student_id']}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
