---
name: github-release-creation
description: Create a GitHub release for the PROMPT 2.0 repository. Use this when asked to create, draft, or publish a GitHub release.
---

Follow these steps strictly to create a GitHub release for the PROMPT 2.0 repository.

## Step 1 — Determine the version number

Read `clients/core/package.json` and extract the `"version"` field (e.g. `2.9.0`).
Prefix it with `v` to form the tag name (e.g. `v2.9.0`).

**Ask the user to confirm:**
> The version detected from `clients/core/package.json` is **v{version}**. Is this the correct version tag for the release?

Wait for confirmation before continuing. If the user provides a different version, use that one.

## Step 2 — Collect context

Run all of these commands to gather the information you need:

```bash
# Find the last published release tag
gh release list --repo prompt-edu/prompt --limit 5

# Get all merged PRs since the last release tag (replace LAST_TAG with the actual tag)
git log LAST_TAG..main --oneline --merges

# Get detailed PR info for the changelog — list merged PRs with title, number, author and URL
gh pr list --repo prompt-edu/prompt --state merged --base main --limit 100 \
  --json number,title,author,mergedAt,url \
  --jq 'sort_by(.mergedAt) | reverse | .[]'
```

Filter the PR list to only PRs merged **after** the last release date. Cross-reference with the git log to be precise.
Exclude dependency-only PRs (Renovate / Dependabot) from the main sections — group them separately under **Dependency Updates**.
Also exclude version bump PRs (e.g. "Bump version to X.Y.Z") from the main change sections.

## Step 3 — Ask for custom additions and breaking changes

Ask the user two questions:

1. **Custom additions:** Are there any custom items, highlights, or context you would like to add to the release notes?
2. **Breaking changes:** Are there any breaking changes in this release that need to be called out? (e.g. changed environment variables, removed APIs, migration steps required)

Wait for both answers before drafting the release body.

## Step 4 — Draft the release body

Compose the release notes using the following structure. Use clear, friendly language — write for a mixed audience of developers, instructors, and students.

```markdown
## Summary

{2–4 sentences describing the main themes of this release. Highlight the most impactful changes for users and developers. Keep it high-level and welcoming.}

{If breaking changes were provided, add a clearly visible warning block here:}
> [!WARNING]
> **Breaking Changes**
> {List each breaking change on its own line with a brief explanation and any migration steps.}

---

## What's New

### 🎓 Student Facing
{Changes that directly affect the student experience — application submission, certificate pages, self team allocation, student-visible course phases, etc.}
- PR title summary ([#NNN](PR_URL)) by @author

### 👩‍🏫 Instructor Facing
{Changes for course instructors and editors — interview management, team allocation & matching, assessment, intro course, course phase configuration, etc.}
- PR title summary ([#NNN](PR_URL)) by @author

### 🔧 Admin Facing
{Changes for platform admins — core system, authentication, multi-tenancy, deployment, infrastructure.}
- PR title summary ([#NNN](PR_URL)) by @author

### 🛠️ Development & Internal
{Developer-facing changes — new APIs, database migrations, testing improvements, documentation, performance, refactoring, build tooling, CI/CD.}
- PR title summary ([#NNN](PR_URL)) by @author

---

## Dependency Updates
{List all Renovate/Dependabot dependency PRs, condensed.}
- Dependency description ([#NNN](PR_URL))

---

**Full Changelog**: https://github.com/prompt-edu/prompt/compare/LAST_TAG...NEW_TAG
```

### Categorisation guide

Use the PR title tags (backtick-wrapped prefixes like `` `Application` ``, `` `Interview` ``) to assign PRs to sections:

| Section | Tags |
|---|---|
| Student Facing | `Application`, `Certificate`, `Self Team Allocation` |
| Instructor Facing | `Interview`, `Team Allocation`, `Matching`, `Assessment`, `Intro Course`, `Templating`, `New Phase`, `Course Phases` |
| Admin Facing | `Core`, `Authentication`, `Infrastructure`, `Deployment` |
| Development & Internal | `General`, `Database`, `API`, `Testing`, `Documentation`, `Module Federation`, `Performance`, `UI/UX`, `Security` |

If a PR tag is ambiguous or missing, use your judgement based on the PR title and content.
Omit any section that has no entries.

If custom additions were provided by the user, incorporate them naturally into the relevant section or add a highlighted callout in the Summary.

## Step 5 — User approval

Present the full draft release body to the user and ask:
> Does this release text look good? Please confirm to publish, suggest edits, or ask me to change anything before I create the draft release.

Do **not** run `gh release create` until the user explicitly approves.

## Step 6 — Create the draft release

Once the user approves, run:

```bash
gh release create NEW_TAG \
  --repo prompt-edu/prompt \
  --title "NEW_TAG" \
  --notes "APPROVED_BODY" \
  --draft \
  --target main
```

- Always use `--draft` so the user can review on GitHub before publishing.
- Always use `--target main`.
- Replace `NEW_TAG` with the confirmed version tag (e.g. `v2.9.0`).

After the command succeeds, confirm the draft release URL to the user.
