#!/usr/bin/env bash
#
# pr-unresolved.sh — list unresolved review threads (comments + inline
# suggestions) on a GitHub pull request, as JSON.
#
# Usage:
#   pr-unresolved.sh [PR_NUMBER]
#
# If PR_NUMBER is omitted, the PR associated with the current branch is used.
# Requires: gh (authenticated), jq.

set -euo pipefail

# --- preflight ---------------------------------------------------------------
for cmd in gh jq; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "error: '$cmd' is required but not installed" >&2
    exit 1
  fi
done

# --- resolve repo + PR number ------------------------------------------------
owner=$(gh repo view --json owner --jq .owner.login)
repo=$(gh repo view --json name --jq .name)

if [[ $# -ge 1 ]]; then
  pr="$1"
else
  if ! pr=$(gh pr view --json number --jq .number 2>/dev/null); then
    echo "error: no PR number given and no PR found for the current branch" >&2
    exit 1
  fi
fi

# --- query -------------------------------------------------------------------
# Paginates through all review threads. Emits a JSON array of unresolved
# threads, each with location, status flags, and the full comment chain.
gh api graphql --paginate -f query='
query($owner: String!, $repo: String!, $pr: Int!, $endCursor: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      reviewThreads(first: 100, after: $endCursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          isResolved
          isOutdated
          path
          line
          originalLine
          comments(first: 100) {
            nodes {
              author { login }
              body
              url
              createdAt
            }
          }
        }
      }
    }
  }
}' -F owner="$owner" -F repo="$repo" -F pr="$pr" \
  --jq '.data.repository.pullRequest.reviewThreads.nodes[]
        | select(.isResolved == false and .isOutdated == false)
        | {
            path,
            line: (.line // .originalLine),
            isOutdated,
            comments: [
              .comments.nodes[]
              | { author: .author.login, body, url, createdAt }
            ]
          }' \
  | jq -s '.'
