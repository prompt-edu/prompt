// All-out "try to break it" scenario. Bounded so it stresses the SERVERS, not
// your laptop. Three parallel attack lanes:
//   1. flood    - very high concurrency on DB-heavy read endpoints (pool saturation)
//   2. bigbody  - multi-MB JSON bodies to a write endpoint (body-size handling)
//   3. nested   - deeply-nested JSON (parser blowup / stack)
// A true slowloris (partial-send) lane lives in the Python fuzzer (raw sockets);
// k6 can't hold partial sends. Abort guard stops the run if it melts down.

import http from "k6/http";
import { Counter, Trend, Rate } from "k6/metrics";
import {
  catalog,
  services,
  buildUrl,
  headersForRoles,
  tags,
} from "./lib/catalog.js";

const INTENSITY = (__ENV.INTENSITY || "medium").toLowerCase();
const PRESET = {
  gentle: { floodVus: 50, bigMB: 1, nestDepth: 2000, dur: "30s" },
  medium: { floodVus: 200, bigMB: 5, nestDepth: 8000, dur: "45s" },
  brutal: { floodVus: 600, bigMB: 20, nestDepth: 40000, dur: "60s" },
}[INTENSITY] || { floodVus: 200, bigMB: 5, nestDepth: 8000, dur: "45s" };

export const reqs = new Counter("reqs");
export const serverErrors = new Counter("server_errors");
export const epLatency = new Trend("endpoint_latency", true);
export const fatal = new Rate("fatal_rate");

// heaviest read endpoints (full-table / join-ish) for pool saturation
const HEAVY = catalog.filter(
  (e) =>
    e.method === "GET" &&
    (e.category === "read" || e.category === "public") &&
    (e.path.endsWith("/students/with-courses") ||
      e.path.endsWith("/students/") ||
      e.path.endsWith("/courses/") ||
      e.path.includes("/participations") ||
      e.path.includes("/student-assessment") ||
      e.path.includes("/allocation"))
);

const CORE = services.services.core.base_url;
const ADMIN_H = () => headersForRoles(["PromptAdmin"]);

// oversized body built once at init
function bigPayload(mb) {
  const chunk = "A".repeat(1024 * 1024);
  let s = "";
  for (let i = 0; i < mb; i++) s += chunk;
  return JSON.stringify({ name: s, semesterTag: "x", courseType: "lecture", restrictedData: {}, studentReadableData: {} });
}
function nestedPayload(depth) {
  // {"a":{"a":{"a": ... }}} as a string to avoid building a huge JS object
  let open = "",
    close = "";
  for (let i = 0; i < depth; i++) {
    open += '{"a":';
    close += "}";
  }
  return open + "1" + close;
}
const BIG = bigPayload(PRESET.bigMB);
const NESTED = nestedPayload(PRESET.nestDepth);

export const options = {
  thresholds: {
    // abort only on a true 5xx/timeout meltdown, not on 4xx
    fatal_rate: [{ threshold: "rate<0.75", abortOnFail: true, delayAbortEval: "20s" }],
  },
  scenarios: {
    flood: {
      executor: "constant-vus",
      vus: PRESET.floodVus,
      duration: PRESET.dur,
      exec: "flood",
    },
    bigbody: {
      executor: "constant-vus",
      vus: 5,
      duration: PRESET.dur,
      exec: "bigbody",
    },
    nested: {
      executor: "constant-vus",
      vus: 5,
      duration: PRESET.dur,
      exec: "nested",
    },
  },
};

function record(ep, res, lane) {
  const t = tags(ep, "exhaustion");
  t.lane = lane;
  t.status = String(res.status);
  reqs.add(1, t);
  epLatency.add(res.timings.duration, { endpoint: ep.id, scenario: "exhaustion" });
  const bad = res.status >= 500 || res.status === 0;
  fatal.add(bad);
  if (bad) serverErrors.add(1, t);
}

export function flood() {
  const pool = HEAVY.length ? HEAVY : catalog.filter((e) => e.method === "GET");
  const ep = pool[(__VU + __ITER) % pool.length];
  const res = http.get(buildUrl(ep), { headers: headersForRoles(ep.roles), tags: Object.assign(tags(ep, "exhaustion"), { lane: "flood" }), timeout: "30s" });
  record(ep, res, "flood");
}

export function bigbody() {
  const synthetic = { id: "core.POST.bigbody", service: "core", method: "POST", path: "/api/courses/", category: "write", roles: ["PromptAdmin"] };
  const res = http.post(`${CORE}/api/courses/`, BIG, { headers: ADMIN_H(), tags: { endpoint: synthetic.id, service: "core", scenario: "exhaustion", lane: "bigbody" }, timeout: "60s" });
  record(synthetic, res, "bigbody");
}

export function nested() {
  const synthetic = { id: "core.POST.nestedbody", service: "core", method: "POST", path: "/api/courses/", category: "write", roles: ["PromptAdmin"] };
  const res = http.post(`${CORE}/api/courses/`, NESTED, { headers: ADMIN_H(), tags: { endpoint: synthetic.id, service: "core", scenario: "exhaustion", lane: "nested" }, timeout: "60s" });
  record(synthetic, res, "nested");
}
