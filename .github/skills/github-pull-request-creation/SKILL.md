---
name: github-pull-request-creation
description: Guide to creating a GitHub Pull Request. Use this when asked to create a PR for your changes.
---

To create a GitHub Pull Request using the GH CLI, follow these rules strictly:

1. Preflight first.
   - Resolve the current branch name and ensure it is pushed.
   - Check whether a PR already exists for this head branch before creating one: `GH_PAGER=cat PAGER=cat gh pr list --head <current-branch> --json number,title,url,state`.
   - If a PR already exists, do not run `gh pr create` again. Return the existing PR URL.
2. Read the PR template at `.github/pull_request_template.md` in the repo.
3. Fill in every section of the template using your knowledge of the changes on the current branch (read commits, diffs, and changed files as needed).
4. Preserve the template structure and text.
   - Do not delete template headings, labels, checklist items, table rows, or helper text.
   - Do not add extra Markdown headings or sections beyond what the template contains.
   - Keep existing wording from the template and fill it out rather than rewriting/removing it.
5. Emoji policy.
   - Do not add emojis anywhere in PR title or newly written body text.
   - If the template already contains emojis in headings, keep them exactly as-is (do not remove or alter them).
6. Backticks are mandatory.
   - The feature tag in the PR title must be enclosed in backticks.
   - Use backticks for all literal commands, file paths, branch names, and endpoint/path literals you mention in the PR body.
7. Do not create files outside the repository (no `/tmp/pr_body.md` or similar).
   - Preferred: write the PR body into a temporary file inside the repo (for example `.pr_body_tmp.md`) via normal file editing tools.
   - Delete the temporary file immediately after the PR is created, before ending your turn.
8. Shell safety (must avoid heredoc/prompt lockups).
   - Never use multiline heredocs in the shared foreground shell for PR body creation.
   - Prefer a single command invocation that references an already-written in-repo body file.
   - Use `GH_PAGER=cat PAGER=cat` for GH CLI commands to avoid pager interaction.
   - If terminal state looks stuck (continuation prompt), stop and recover shell state before running more commands.
9. Use `gh pr create` with appropriate `--title`, `--body` (or `--body-file`), `--base main`, and `--head <current-branch>` flags.
10. Derive the PR title using this exact format: `` `Feature Tag` ``: Short verbal description

- Choose the most fitting tag from: `Core`, `Course Phases`, `Authentication`, `Module Federation`, `Interview`, `Team Allocation`, `Matching`, `Assessment`, `Certificate`, `Intro Course`, `Self Team Allocation`, `Templating`, `New Phase`, `Database`, `API`, `Testing`, `Documentation`, `Infrastructure`, `Deployment`, `UI/UX`, `Performance`, `General`. Use `General` if none fit.
- The tag must be enclosed in backticks (e.g. `` `Core` ``) — this is mandatory.
- After the colon (outside backticks), write a short description in plain, non-technical language suitable for release notes.
- Sentence case only (capitalize the first word, nothing else).
- Use the verb "Fix" for bugfixes; use other action verbs (Add, Allow, Show, Enable, …) for improvements.
- Do not repeat the feature tag in the description if avoidable.
- Example: `` `Core` ``: Allow instructors to create and manage course phases

11. Check the relevant checklist items that apply (replace `[ ]` with `[x]` for items that are confirmed true, e.g. "Tested locally" if the description says so).
12. After the PR is successfully created, confirm the PR URL to the user.
