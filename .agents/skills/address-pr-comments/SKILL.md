---
name: address-pr-comments
description: Address unresolved review comments on the current branch's GitHub PR — fetch the open threads, fix the code, and push. Use when asked to handle, resolve, or respond to PR review comments/feedback.
---

Address the unresolved review threads on the current branch's PR. Assume you are on the PR's branch,
`gh` is authenticated, and `jq` is installed. This is a simple task — keep shell use to the two
commands below.

1. List the unresolved threads as JSON (path, line, and the full comment chain) with one command:

   ```bash
   ./scripts/pr-unresolved.sh
   ```

2. For each thread, make the requested change in the code (follow the relevant `.claude/rules/`).
   Group the fixes into clear commits.

3. Push the fixes with one command:

   ```bash
   git push
   ```

Optionally reply to a specific thread to confirm what changed (`gh pr comment` or `gh api`), but
don't add ceremony — the push plus the code changes are the substance. Re-run
`./scripts/pr-unresolved.sh` if you need to confirm nothing was missed.
