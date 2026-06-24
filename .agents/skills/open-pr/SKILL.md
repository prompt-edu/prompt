---
name: open-pr
description: Open a GitHub pull request for the current branch using this repo's PR template and the gh CLI. Use when asked to open, create, or raise a PR.
---

Open a PR for the current branch. Assume the working branch is correct, `gh` is installed and
authenticated, and you are in this repo. This is a simple task — keep it to one or two commands.

1. Compose the PR locally (no shell commands needed):
   - **Title** in the format `` `Tag`: short description `` — backticked feature tag, then a plain
     sentence-case summary suitable for release notes. Use "Fix" for bugfixes, action verbs (Add,
     Allow, Show, Enable…) otherwise. Tags: `Core`, `Course Phases`, `Authentication`,
     `Module Federation`, `Interview`, `Team Allocation`, `Matching`, `Assessment`, `Certificate`,
     `Intro Course`, `Self Team Allocation`, `Templating`, `New Phase`, `Database`, `API`, `Testing`,
     `Documentation`, `Infrastructure`, `Deployment`, `UI/UX`, `Performance`, `General` (use `General`
     if none fit).
   - **Body:** fill in every section of `.github/pull_request_template.md` from the branch's commits
     and diff. Keep the template's headings, checklists, and helper text; check the items that are
     true. Write the filled body to a temporary file in the repo (e.g. `.pr-body.md`) with your
     editor.

2. Create the PR with a single command (pushes the branch if needed):

   ```bash
   git push -u origin HEAD 2>/dev/null; gh pr create --base main --title "<title>" --body-file .pr-body.md
   ```

3. Delete `.pr-body.md` and report the PR URL.
