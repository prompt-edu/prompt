---
name: open-pr
description: Open a GitHub pull request for the current branch using this repo's PR template and the gh CLI. Use when asked to open, create, or raise a PR.
---

Open a PR for the current branch. Assume the working branch is correct, `gh` is installed and
authenticated, and you are in this repo. This is a simple task — keep it to one or two commands and
use `GH_PAGER=cat` on `gh` calls to avoid pager hangs.

1. Preflight (one command): if a PR already exists for this branch, don't create another — just
   return its URL.

   ```bash
   GH_PAGER=cat gh pr list --head "$(git branch --show-current)" --json number,url,state
   ```

2. Compose the PR locally (no shell commands needed):
   - **Title** — `` `Tag`: short description ``: a backticked feature tag, then a plain sentence-case
     summary suitable for release notes (capitalize only the first word). Use "Fix" for bugfixes,
     other action verbs (Add, Allow, Show, Enable…) otherwise; don't repeat the tag in the
     description. Tags: `Core`, `Course Phases`, `Authentication`, `Module Federation`, `Interview`,
     `Team Allocation`, `Matching`, `Assessment`, `Certificate`, `Intro Course`,
     `Self Team Allocation`, `Templating`, `New Phase`, `Database`, `API`, `Testing`,
     `Documentation`, `Infrastructure`, `Deployment`, `UI/UX`, `Performance`, `General` (use
     `General` if none fit).
   - **Body** — fill in every section of `.github/pull_request_template.md` from the branch's commits
     and diff. Keep all headings, helper text, and checklist items; check the boxes that are true
     (`[x]`). Use backticks for literal commands, paths, and branch names. Don't add emojis to the
     title or new body text (keep any emojis already in the template). Write the body to a temp file
     in the repo (e.g. `.pr_body_tmp.md`) with your editor — never outside the repo.

3. Create the PR with a single command (pushes the branch if needed):

   ```bash
   git push -u origin HEAD 2>/dev/null; GH_PAGER=cat gh pr create --base main --head "$(git branch --show-current)" --title "<title>" --body-file .pr_body_tmp.md
   ```

4. Delete `.pr_body_tmp.md` and report the PR URL.
