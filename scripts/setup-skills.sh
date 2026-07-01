#!/usr/bin/env bash
# Regenerate .claude/skills symlinks from the canonical skills in .agents/skills.
#
# Skill content is tracked in .agents/skills/ (the cross-tool source of truth).
# Claude Code discovers skills from .claude/skills/, so we create one symlink per
# skill (not a whole-directory link, which would let Claude write .system/ files
# back into .agents/ and pollute it for other tools). .claude/skills/ is gitignored.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
src="$repo_root/.agents/skills"
dst="$repo_root/.claude/skills"

if [ ! -d "$src" ]; then
  echo "No .agents/skills directory found at $src" >&2
  exit 1
fi

mkdir -p "$dst"

# Remove existing skill symlinks so links for deleted skills don't linger and
# .claude/skills/ stays in sync with the canonical .agents/skills/ set. Also clear
# any stale real directory left by an older whole-directory setup, otherwise the
# `ln -sfn` below would nest a dangling link inside it instead of replacing it.
find "$dst" -mindepth 1 -maxdepth 1 \( -type l -o -type d \) -exec rm -rf {} +

count=0
for skill_dir in "$src"/*/; do
  [ -d "$skill_dir" ] || continue
  name="$(basename "$skill_dir")"
  ln -sfn "../../.agents/skills/$name" "$dst/$name"
  echo "linked .claude/skills/$name -> ../../.agents/skills/$name"
  count=$((count + 1))
done

echo "Done. $count skill symlink(s) regenerated in .claude/skills/."
