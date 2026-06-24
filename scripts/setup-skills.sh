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

count=0
for skill_dir in "$src"/*/; do
  [ -d "$skill_dir" ] || continue
  name="$(basename "$skill_dir")"
  ln -sfn "../../.agents/skills/$name" "$dst/$name"
  echo "linked .claude/skills/$name -> ../../.agents/skills/$name"
  count=$((count + 1))
done

echo "Done. $count skill symlink(s) regenerated in .claude/skills/."
