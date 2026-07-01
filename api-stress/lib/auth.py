#!/usr/bin/env python3
"""Mint and cache Keycloak access tokens per test user (password grant).

Tokens are obtained ONCE per role and cached to reports/<run>/tokens.json so that
neither k6 nor the fuzzer hammer Keycloak during the load phase (Keycloak is
explicitly out of scope as a load target). Course-scoped role tokens MUST be
minted AFTER fixtures.py has added the users to the new course's Keycloak groups,
otherwise the token carries no course role and every course-scoped call 403s.

Usable both as a library (fetch_token) and a CLI (writes tokens.json).
"""
from __future__ import annotations

import json
import pathlib
import sys
import time

import httpx

HERE = pathlib.Path(__file__).parent
SERVICES = json.load(open(HERE / "services.json"))
KC = SERVICES["keycloak"]
TOKEN_URL = KC["base_url"] + KC["token_path"]
CLIENT_ID = "prompt-client"  # public client, directAccessGrantsEnabled

# Seeded realm users (keycloakConfig.json). password == username for all.
USERS = {
    "admin": "admin",
    "lecturer": "lecturer",
    "course-lecturer": "course-lecturer",
    "course-editor": "course-editor",
    "student": "student",
}


def fetch_token(username: str, password: str, timeout: float = 10.0) -> dict:
    data = {
        "grant_type": "password",
        "client_id": CLIENT_ID,
        "username": username,
        "password": password,
        "scope": "openid",
    }
    last = None
    for attempt in range(5):
        try:
            r = httpx.post(TOKEN_URL, data=data, timeout=timeout)
            if r.status_code == 200:
                return r.json()
            last = f"HTTP {r.status_code}: {r.text[:200]}"
        except Exception as e:  # noqa: BLE001
            last = repr(e)
        time.sleep(1.5 * (attempt + 1))
    raise RuntimeError(f"could not get token for {username}: {last}")


def main() -> int:
    out_path = pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 else HERE / "tokens.json"
    tokens = {}
    errors = {}
    for user, pw in USERS.items():
        try:
            tokens[user] = fetch_token(user, pw)["access_token"]
            print(f"  token ok: {user}")
        except Exception as e:  # noqa: BLE001
            errors[user] = str(e)
            print(f"  token FAIL: {user}: {e}", file=sys.stderr)
    json.dump(
        {"tokens": tokens, "errors": errors, "token_url": TOKEN_URL, "client_id": CLIENT_ID},
        open(out_path, "w"),
        indent=2,
    )
    print(f"wrote {len(tokens)} tokens -> {out_path}")
    return 0 if tokens else 1


if __name__ == "__main__":
    sys.exit(main())
