#!/usr/bin/env python3
"""Merge per-service partial route files into the canonical endpoints.json.

The catalog is the single source of truth consumed by both the k6 scenarios and
the Python fuzzer. Run after (re)generating any partial_<service>.json.
"""
import json
import pathlib
import sys

HERE = pathlib.Path(__file__).parent
PARTIALS = sorted(HERE.glob("partial_*.json"))

VALID_METHODS = {"GET", "POST", "PUT", "DELETE", "PATCH"}
VALID_CATEGORY = {"read", "write", "destructive", "public"}


def main() -> int:
    out = []
    seen_ids = set()
    problems = []
    for p in PARTIALS:
        try:
            rows = json.load(open(p))
        except Exception as e:  # noqa: BLE001
            problems.append(f"{p.name}: invalid JSON ({e})")
            continue
        for r in rows:
            rid = r.get("id")
            if not rid or rid in seen_ids:
                # de-dup / repair missing ids deterministically
                rid = f"{r.get('service','?')}.{r.get('method','?')}.{len(out)}"
                r["id"] = rid
            seen_ids.add(rid)
            m = r.get("method", "").upper()
            if m not in VALID_METHODS:
                problems.append(f"{rid}: bad method {m!r}")
            if r.get("category") not in VALID_CATEGORY:
                # infer
                r["category"] = (
                    "public" if r.get("roles") == ["public"]
                    else "destructive" if m == "DELETE"
                    else "read" if m == "GET"
                    else "write"
                )
            r.setdefault("roles", ["public"])
            r.setdefault("pathParams", [])
            r.setdefault("bodyType", None)
            out.append(r)

    out.sort(key=lambda r: (r.get("service", ""), r.get("path", ""), r.get("method", "")))
    catalog = {
        "_comment": "CANONICAL endpoint catalog (auto-merged from partial_*.json). "
        "Single source of truth for the stress/fuzz suite. Regenerate via merge_catalog.py.",
        "count": len(out),
        "endpoints": out,
    }
    dest = HERE / "endpoints.json"
    json.dump(catalog, open(dest, "w"), indent=2)
    by_service = {}
    for r in out:
        by_service[r["service"]] = by_service.get(r["service"], 0) + 1
    print(f"merged {len(out)} routes -> {dest}")
    for s, n in sorted(by_service.items()):
        print(f"  {s:24s} {n}")
    if problems:
        print("\nPROBLEMS:", file=sys.stderr)
        for pr in problems:
            print("  " + pr, file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
