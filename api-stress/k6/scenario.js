// Parametrized k6 driver for smoke / load / spike / soak.
//
//   SCENARIO=smoke  -> 1 VU walks EVERY endpoint once (reachability + baseline).
//   SCENARIO=load   -> ramping VUs over read+public endpoints; reveals the knee.
//   SCENARIO=spike  -> sudden burst then recovery.
//   SCENARIO=soak   -> sustained moderate load (slow leaks / drift).
//
// Per-request we emit: built-in http_req_duration tagged {endpoint,scenario,...}
// plus a custom `reqs` counter tagged with the response status, so the Python
// report can aggregate latency percentiles AND status mix per endpoint from the
// k6 JSON output. Thresholds abort a run if it melts down (protects the host).

import http from "k6/http";
import { Counter, Trend, Rate } from "k6/metrics";
import exec from "k6/execution";
import {
  catalog,
  buildUrl,
  genericBody,
  headersForRoles,
  tags,
} from "./lib/catalog.js";

const SCENARIO = __ENV.SCENARIO || "smoke";
const INTENSITY = (__ENV.INTENSITY || "medium").toLowerCase();

// intensity presets (peak VUs for load/spike, VUs for soak)
const PRESET = {
  gentle: { loadPeak: 30, spikePeak: 60, soakVus: 10, soakDur: "1m" },
  medium: { loadPeak: 120, spikePeak: 250, soakVus: 40, soakDur: "2m" },
  brutal: { loadPeak: 400, spikePeak: 800, soakVus: 120, soakDur: "4m" },
}[INTENSITY] || { loadPeak: 120, spikePeak: 250, soakVus: 40, soakDur: "2m" };

export const reqs = new Counter("reqs");
export const serverErrors = new Counter("server_errors");
export const epLatency = new Trend("endpoint_latency", true);
// "fatal" = 5xx or connection error. We abort the run on FATAL meltdown only -
// NOT on 4xx (many catalog reads legitimately 404 on non-seeded sub-resources,
// and that must not be mistaken for the server falling over).
export const fatal = new Rate("fatal_rate");

// read + public GET endpoints: safe to hammer, meaningful DB load
const READ_POOL = catalog.filter(
  (e) => e.method === "GET" && (e.category === "read" || e.category === "public")
);

function buildOptions() {
  // abort the whole run only on genuine meltdown: >50% of requests returning
  // 5xx/conn-error for 15s straight. 4xx is expected and never aborts.
  const guard = {
    thresholds: {
      fatal_rate: [{ threshold: "rate<0.50", abortOnFail: true, delayAbortEval: "15s" }],
    },
  };
  if (SCENARIO === "smoke") {
    return Object.assign(guard, {
      scenarios: {
        smoke: { executor: "per-vu-iterations", vus: 1, iterations: 1, maxDuration: "10m" },
      },
    });
  }
  if (SCENARIO === "load") {
    return Object.assign(guard, {
      scenarios: {
        load: {
          executor: "ramping-vus",
          startVUs: 1,
          stages: [
            { duration: "15s", target: 10 },
            { duration: "20s", target: 50 },
            { duration: "20s", target: Math.round(PRESET.loadPeak / 2) },
            { duration: "25s", target: PRESET.loadPeak },
            { duration: "15s", target: PRESET.loadPeak },
            { duration: "10s", target: 0 },
          ],
          gracefulStop: "10s",
        },
      },
    });
  }
  if (SCENARIO === "spike") {
    return Object.assign(guard, {
      scenarios: {
        spike: {
          executor: "ramping-vus",
          startVUs: 1,
          stages: [
            { duration: "10s", target: 10 },
            { duration: "5s", target: PRESET.spikePeak }, // sudden burst
            { duration: "20s", target: PRESET.spikePeak },
            { duration: "5s", target: 5 }, // crash back down
            { duration: "15s", target: 5 }, // observe recovery
            { duration: "5s", target: 0 },
          ],
          gracefulStop: "10s",
        },
      },
    });
  }
  // soak
  return Object.assign(guard, {
    scenarios: {
      soak: {
        executor: "constant-vus",
        vus: PRESET.soakVus,
        duration: PRESET.soakDur,
      },
    },
  });
}

export const options = buildOptions();

function fire(ep) {
  const url = buildUrl(ep);
  const body = genericBody(ep);
  const params = { headers: headersForRoles(ep.roles), tags: tags(ep, SCENARIO), timeout: "30s" };
  let res;
  if (body === null) res = http.request(ep.method, url, null, params);
  else res = http.request(ep.method, url, body, params);

  const t = tags(ep, SCENARIO);
  t.status = String(res.status);
  reqs.add(1, t);
  epLatency.add(res.timings.duration, { endpoint: ep.id, scenario: SCENARIO });
  const bad = res.status >= 500 || res.status === 0;
  fatal.add(bad);
  if (bad) serverErrors.add(1, t);
  return res;
}

export default function () {
  if (SCENARIO === "smoke") {
    // one VU, one iteration: walk the whole catalog deterministically
    for (let i = 0; i < catalog.length; i++) fire(catalog[i]);
    return;
  }
  // load/spike/soak: hit a random read/public endpoint
  const pool = READ_POOL.length ? READ_POOL : catalog;
  const ep = pool[exec.scenario.iterationInTest % pool.length];
  fire(ep);
}
