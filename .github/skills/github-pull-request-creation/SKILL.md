---
name: github-pull-request-creation
description: Guide to creating a GitHub Pull Request. Use this when asked to create a PR for your changes.
---

To create a GitHub Pull Request using the GH CLI, follow these rules strictly:

1. Read the PR template at `.github/pull_request_template.md` in the repo.
2. Fill in every section of the template using your knowledge of the changes on the current branch (read commits, diffs, and changed files as needed).
3. Do not add any extra Markdown headings or sections beyond what the template contains.
4. Do not create any files outside the repository (no `/tmp/pr_body.md` or similar). Instead, write the filled-out body inline using the `--body` flag of `gh pr create`, or use a temporary file inside the repo (e.g. `.pr_body_tmp.md`) and delete it immediately after the PR is created — before ending your turn.
5. Use `gh pr create` with appropriate `--title`, `--body` (or `--body-file`), `--base main`, and `--head <current-branch>` flags.
6. Derive the PR title using this exact format: `` `Feature Tag` ``: Short verbal description
   - Choose the most fitting tag from: `Core`, `Course Phases`, `Authentication`, `Module Federation`, `Interview`, `Team Allocation`, `Matching`, `Assessment`, `Certificate`, `Intro Course`, `Self Team Allocation`, `Templating`, `New Phase`, `Database`, `API`, `Testing`, `Documentation`, `Infrastructure`, `Deployment`, `UI/UX`, `Performance`, `General`. Use `General` if none fit.
   - The tag must be enclosed in backticks (e.g. `` `Core` ``) — this is mandatory.
   - After the colon (outside backticks), write a short description in plain, non-technical language suitable for release notes.
   - Sentence case only (capitalize the first word, nothing else).
   - Use the verb "Fix" for bugfixes; use other action verbs (Add, Allow, Show, Enable, …) for improvements.
   - Do not repeat the feature tag in the description if avoidable.
   - Example: `` `Core` ``: Allow instructors to create and manage course phases
7. Check the relevant checklist items that apply (replace `[ ]` with `[x]` for items that are confirmed true, e.g. "Tested locally" if the description says so).
8. After the PR is successfully created, confirm the PR URL to the user.
