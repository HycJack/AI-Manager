#!/bin/bash
# Migrate skills to ~/.aimanager/skills/
set -e

DST="$HOME/.aimanager/skills"
mkdir -p "$DST"
count=0

# 1. ~/.agents/skills/ (universal skills)
if [ -d "$HOME/.agents/skills" ]; then
  echo "Migrating ~/.agents/skills/ → ~/.aimanager/skills/ ..."
  for dir in "$HOME/.agents/skills"/*/; do
    [ -d "$dir" ] || continue
    name=$(basename "$dir")
    if [ ! -d "$DST/$name" ]; then
      cp -R "$dir" "$DST/$name"
      echo "  ✓ $name"
      count=$((count + 1))
    else
      echo "  skip $name (already exists)"
    fi
  done
else
  echo "No ~/.agents/skills/ found."
fi

# 2. ~/Library/Application Support/AIManager/library/skills/ (macOS Library)
MAC_OS_DIR="$HOME/Library/Application Support/AIManager/library/skills"
if [ -d "$MAC_OS_DIR" ]; then
  echo ""
  echo "Migrating ~/Library/Application Support/AIManager/library/skills/ → ~/.aimanager/skills/ ..."
  for dir in "$MAC_OS_DIR"/*/; do
    [ -d "$dir" ] || continue
    name=$(basename "$dir")
    if [ ! -d "$DST/$name" ]; then
      cp -R "$dir" "$DST/$name"
      echo "  ✓ $name"
      count=$((count + 1))
    else
      echo "  skip $name (already exists)"
    fi
  done
else
  echo ""
  echo "No ~/Library/Application Support/AIManager/library/skills/ found."
fi

# 3. ~/.ai-manager/library/skills/ (fallback)
if [ -d "$HOME/.ai-manager/library/skills" ]; then
  echo ""
  echo "Migrating ~/.ai-manager/library/skills/ → ~/.aimanager/skills/ ..."
  for dir in "$HOME/.ai-manager/library/skills"/*/; do
    [ -d "$dir" ] || continue
    name=$(basename "$dir")
    if [ ! -d "$DST/$name" ]; then
      cp -R "$dir" "$DST/$name"
      echo "  ✓ $name"
      count=$((count + 1))
    else
      echo "  skip $name (already exists)"
    fi
  done
else
  echo ""
  echo "No ~/.ai-manager/library/skills/ found."
fi

# 4. Copy registry.json (check both locations)
for reg in "$MAC_OS_DIR/../registry.json" "$HOME/.ai-manager/library/registry.json"; do
  if [ -f "$reg" ]; then
    cp "$reg" "$HOME/.aimanager/registry.json"
    echo ""
    echo "  ✓ registry.json copied"
    break
  fi
done

echo ""
echo "Done. Migrated $count skills to $DST."
echo "Old directories are preserved. Remove manually if desired."
