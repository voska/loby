#!/bin/sh
# loby SKILL.md installer for Claude Code / Codex / Factory Droid.
#
# Drops the SKILL.md + references into the right directory so an AI agent
# can discover loby on its next invocation.
set -eu

REPO="https://raw.githubusercontent.com/voska/loby/main/skills/loby"

# Pick the agent harness dir, in this order: explicit env > Claude Code > Codex
target="${LOBY_SKILL_DIR:-}"
if [ -z "$target" ]; then
  if [ -d "$HOME/.claude" ]; then
    target="$HOME/.claude/skills/loby"
  elif [ -d "$HOME/.agents" ]; then
    target="$HOME/.agents/skills/loby"
  else
    target="$HOME/.claude/skills/loby"
  fi
fi

mkdir -p "$target/references"

fetch() {
  url="$REPO/$1"
  dest="$target/$1"
  printf 'fetching %s\n' "$1"
  curl -fsSL "$url" -o "$dest"
}

fetch SKILL.md
fetch references/COMMANDS.md
fetch references/RECIPES.md
fetch references/RESOURCES.md

printf '\nInstalled loby SKILL.md bundle to %s\n' "$target"
printf 'Ensure the loby binary is installed: brew install voska/tap/loby\n'
