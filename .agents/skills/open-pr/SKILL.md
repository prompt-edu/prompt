---
name: open-pr
description: Open a GitHub pull request for the current branch using this repo's PR template and the gh CLI, with before/after screenshots for UI changes. Use when asked to open, create, or raise a PR, or to add screenshots to one.
---

Open a PR for the current branch. Assume the working branch is correct, `gh` is installed and
authenticated, and you are in this repo. Apart from taking screenshots, this is a simple task —
keep it to one or two commands and use `GH_PAGER=cat` on `gh` calls to avoid pager hangs.

1. Preflight (one command): if a PR already exists for this branch, don't create another — just
   return its URL. When asked to add screenshots to it, run steps 2 and 3 on its current body
   (`GH_PAGER=cat gh pr view --json body --jq .body`) and save it with `gh pr edit` in step 4.

   ```bash
   GH_PAGER=cat gh pr list --head "$(git branch --show-current)" --json number,url,state
   ```

2. Screenshots (visible changes only): a diff that changes what a user sees in `clients/` carries
   the before/after shots the `verify` skill saves under
   `.playwright-mcp/pr-screenshots/<branch>/`. When that folder is missing, holds no before shots,
   or predates the branch's last commit under `clients/`, run `verify` now and take them.
   Refactors and server-only changes carry none.

3. Compose the PR locally (no shell commands needed):
   - **Title** — `` `Tag`: short description ``: a backticked feature tag, then a plain sentence-case
     summary suitable for release notes (capitalize only the first word). Use "Fix" for bugfixes,
     other action verbs (Add, Allow, Show, Enable…) otherwise; don't repeat the tag in the
     description. Tags: `Core`, `Course Phases`, `Authentication`, `Module Federation`, `Interview`,
     `Team Allocation`, `Matching`, `Assessment`, `Certificate`, `Intro Course`,
     `Self Team Allocation`, `Templating`, `New Phase`, `Database`, `API`, `Testing`,
     `Documentation`, `Infrastructure`, `Deployment`, `UI/UX`, `Performance`, `General` (use
     `General` if none fit).
   - **Body** — fill in every section of `.github/pull_request_template.md` from the branch's commits
     and diff. Keep all headings, helper text, and checklist items, except `## Screenshots`,
     which only a PR with screenshots keeps; check the boxes that are true (`[x]`). Use backticks
     for literal commands, paths, and branch names. Don't add emojis to the title or new body text
     (keep any emojis already in the template). Write the body with your editor, never outside
     the repo: to `.pr_body_tmp.md`, or with screenshots to `body.md` in the screenshot folder,
     since `gh` resolves image paths against the body file's folder.
   - **Screenshots** — a table under `## Screenshots` with one row per changed state and width,
     its before and after side by side, each shot referenced as `./<file>.png`. The change cell
     names the state and width, then what changed in a few words. A state only one side has
     shows `_new_` or `_gone_` in the other cell:

     ```markdown
     | Change | Before | After |
     | ------ | ------ | ----- |
     | Participants, desktop: status shown as a badge | ![Participants before, desktop](./01-participants-desktop-before.png) | ![Participants after, desktop](./01-participants-desktop-after.png) |
     ```

4. Create the PR with a single command (pushes the branch if needed):

   ```bash
   git push -u origin HEAD 2>/dev/null; GH_PAGER=cat gh pr create --base main --head "$(git branch --show-current)" --title "<title>" --body-file .pr_body_tmp.md
   ```

   With screenshots, pass `--body-file .playwright-mcp/pr-screenshots/<branch>/body.md` and one
   `--attach .playwright-mcp/pr-screenshots/<branch>/<file>.png` per shot the body references.
   `gh` uploads each and rewrites its `./<file>.png` reference to the upload; it appends a shot the
   body does not reference at the end, so attach only referenced ones. When an upload fails, the
   PR is still created and the command exits non-zero: run `GH_PAGER=cat gh pr edit` with the same
   `--body-file` and every `--attach` again, never a second `gh pr create`. `unknown flag: --attach`
   means `gh` is older than 2.99: `brew upgrade gh`.

5. Delete `.pr_body_tmp.md` or the screenshot folder and report the PR URL.
