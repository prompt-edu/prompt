// Shared k6 helper: load the canonical endpoint catalog + run fixtures/tokens,
// and turn each catalog entry into a concrete, auth'd HTTP request.
//
// Path params are resolved from fixtures.json (the seeded STRESS course A, its
// phases, the enrolled student, etc.). Tokens are chosen as the LEAST-privilege
// role that the endpoint allows, so we exercise the real authz path rather than
// always using admin. Files are read once at init time via k6's open().

const ROOT = __ENV.STRESS_DIR; // absolute path to api-stress/
const RUN = __ENV.RUN_DIR; // absolute path to reports/<run>/

export const services = JSON.parse(open(`${ROOT}/lib/services.json`));
export const catalog = JSON.parse(open(`${ROOT}/catalog/endpoints.json`)).endpoints;
export const fixtures = JSON.parse(open(`${RUN}/fixtures.json`));
export const tokenFile = JSON.parse(open(`${RUN}/tokens.json`));
const TOKENS = tokenFile.tokens || {};

const FALLBACK_UUID = "00000000-0000-0000-0000-000000000000";

// service -> the phase-type name (lowercased) that drives it
const SERVICE_PHASE_TYPE = {
  interview: "interview",
  team_allocation: "team allocation",
  self_team_allocation: "self team allocation",
  assessment: "assessment",
  certificate: "certificate",
  template_server: "template",
  core: "application",
};

// role -> the seeded keycloak username whose token carries it
const ROLE_TO_USER = {
  PromptAdmin: "admin",
  PromptLecturer: "lecturer",
  CourseLecturer: "course-lecturer",
  CourseEditor: "course-editor",
  CourseStudent: "student",
};
// preference order: least privilege first so we test the tightest allowed role
const ROLE_PREF = [
  "CourseStudent",
  "CourseEditor",
  "CourseLecturer",
  "PromptLecturer",
  "PromptAdmin",
];

function phaseIdForService(service) {
  const want = SERVICE_PHASE_TYPE[service] || "application";
  const pa = fixtures.phases_a || {};
  if (pa[want] && pa[want].id) return pa[want].id;
  if (pa["application"] && pa["application"].id) return pa["application"].id;
  const any = Object.values(pa).find((p) => p && p.id);
  return any ? any.id : FALLBACK_UUID;
}

export function tokenForRoles(roles) {
  if (!roles || roles.length === 0 || roles[0] === "public") return null;
  if (roles[0] === "custom") return TOKENS["admin"] || null;
  for (const r of ROLE_PREF) {
    if (roles.includes(r)) {
      const user = ROLE_TO_USER[r];
      if (TOKENS[user]) return TOKENS[user];
    }
  }
  return TOKENS["admin"] || null;
}

// Resolve a single {param} using fixtures + heuristics on the path.
function resolveParam(name, path, service) {
  const ca = (fixtures.course_a && fixtures.course_a.id) || FALLBACK_UUID;
  const student = (fixtures.student && fixtures.student.student_id) || FALLBACK_UUID;
  const part = (fixtures.student && fixtures.student.participation_id) || FALLBACK_UUID;
  switch (name) {
    case "coursePhaseID":
      return phaseIdForService(service);
    case "courseID":
      return ca;
    case "course_participation_id":
    case "courseParticipationID":
      return part;
    case "student-uuid":
      return student;
    case "searchString":
      return "no";
    case "groupName":
      return "Prompt";
    case "uuid":
      if (path.includes("/courses/")) return ca;
      if (path.includes("/course_phases/")) return phaseIdForService("core");
      if (path.includes("/students/")) return student;
      return ca;
    default:
      // teamID, skillID, slotId, schemaID, categoryID, competencyID, assessmentID,
      // evaluationID, tutorID, assignmentId, fileId, docID, note-uuid, tag-uuid...
      return FALLBACK_UUID;
  }
}

export function buildUrl(ep) {
  const svc = services.services[ep.service];
  const base = svc ? svc.base_url : "http://localhost:0";
  let path = ep.path;
  // Reads resolve to REAL fixture ids (true latency baseline). Writes/deletes
  // resolve to a non-existent id so a smoke walk can never corrupt the shared
  // fixtures (e.g. DELETE /courses/{uuid}); a 5xx on a non-existent target is
  // then a genuine error-handling bug, not harness damage. (The Python fuzzer
  // drives real writes separately.)
  const safe = ep.method !== "GET";
  (ep.pathParams || []).forEach((p) => {
    const v = safe ? FALLBACK_UUID : resolveParam(p, ep.path, ep.service);
    path = path.replace(`{${p}}`, v);
  });
  return base + path;
}

// Minimal generic body for write endpoints (load only; the fuzzer does real
// body fuzzing). Empty object exercises bind+validate and measures latency.
export function genericBody(ep) {
  if (ep.method === "GET" || ep.method === "DELETE") return null;
  return JSON.stringify({});
}

export function headersForRoles(roles, extra) {
  const h = Object.assign({ "Content-Type": "application/json" }, extra || {});
  const tok = tokenForRoles(roles);
  if (tok) h["Authorization"] = `Bearer ${tok}`;
  return h;
}

export function tags(ep, scenario) {
  return {
    endpoint: ep.id,
    service: ep.service,
    method: ep.method,
    category: ep.category,
    scenario: scenario,
  };
}
