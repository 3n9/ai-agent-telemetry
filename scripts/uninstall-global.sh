#!/usr/bin/env bash
set -e

GLOBAL_DIR="$HOME/.ai-telemetry"
PROMPTS_DEST="$GLOBAL_DIR/prompts"
BIN_DEST="$GLOBAL_DIR/bin"
BLOCK_NAME="AI TELEMETRY SYSTEM (AUTO-GENERATED)"
BEGIN_MARKER="### BEGIN $BLOCK_NAME"
END_MARKER="### END $BLOCK_NAME"

echo "🗑️  Uninstalling AI Telemetry..."

# 1. REMOVE INJECTED BLOCKS FROM AGENT INSTRUCTION FILES
safe_cleanup() {
  local target="$1"
  local name="$2"

  if [ -f "$target" ] && grep -q "$BEGIN_MARKER" "$target"; then
    echo "🔧 Removing AI Telemetry configuration from $name ($target)..."
    python3 - "$target" "$BEGIN_MARKER" "$END_MARKER" <<'PY'
import pathlib
import sys

target_path = pathlib.Path(sys.argv[1]).expanduser()
begin_marker = sys.argv[2]
end_marker = sys.argv[3]

if target_path.exists():
    content = target_path.read_text()
    start = content.find(begin_marker)
    end = content.find(end_marker, start if start != -1 else 0)
    if start != -1 and end != -1:
        end += len(end_marker)
        content = content[:start] + content[end:]
        content = content.strip("\n")
        if content:
            content += "\n"
        target_path.write_text(content)
PY
    echo "✅ $name configuration cleaned."
  fi
}

safe_cleanup "$HOME/.claude/CLAUDE.md" "Claude Code"
safe_cleanup "$HOME/.copilot/copilot-instructions.md" "Copilot CLI"
safe_cleanup "$HOME/.codex/AGENTS.md" "Codex"
safe_cleanup "$HOME/.aider.conventions.md" "Aider Conventions"

# 2. AGENT: AIDER — remove read reference from ~/.aider.conf.yml
AIDER_CONF="$HOME/.aider.conf.yml"
if [ -f "$AIDER_CONF" ] && grep -q ".aider.conventions.md" "$AIDER_CONF"; then
  echo "🔧 Removing Aider conventions reference from $AIDER_CONF..."
  python3 - "$AIDER_CONF" <<'PY'
import pathlib, re, sys

path = pathlib.Path(sys.argv[1])
content = path.read_text()
# Remove any line containing .aider.conventions.md (the read: entry)
content = re.sub(r'\nread: \[.*?\.aider\.conventions\.md.*?\]', '', content)
content = content.strip("\n")
if content:
    content += "\n"
path.write_text(content)
PY
  echo "✅ Aider conventions reference removed."
fi

# 3. AGENT: GEMINI CLI — remove BeforeAgent hook from settings.json
GEMINI_SETTINGS="$HOME/.gemini/settings.json"
if [ -f "$GEMINI_SETTINGS" ]; then
  echo "🔧 Removing AI Telemetry hook from Gemini CLI settings..."
  python3 - "$GEMINI_SETTINGS" <<'PY'
import json, pathlib, sys

path = pathlib.Path(sys.argv[1])
try:
    data = json.loads(path.read_text())
except (json.JSONDecodeError, FileNotFoundError):
    sys.exit(0)

hooks = data.get("hooks", {}).get("BeforeAgent", [])
data["hooks"]["BeforeAgent"] = [h for h in hooks if h.get("name") != "ai-telemetry-policy"]
path.write_text(json.dumps(data, indent=2) + "\n")
PY
  echo "✅ Gemini CLI hook removed."
fi

# 4. REMOVE SYNCED FILES
if [ -d "$PROMPTS_DEST" ]; then
  rm -rf "$PROMPTS_DEST"
  echo "✅ Global prompts directory removed: $PROMPTS_DEST"
fi

if [ -d "$BIN_DEST" ]; then
  rm -rf "$BIN_DEST"
  echo "✅ Global bin directory removed: $BIN_DEST"
fi

# 5. REMOVE GLOBAL DIR IF EMPTY
if [ -d "$GLOBAL_DIR" ] && [ -z "$(ls -A "$GLOBAL_DIR")" ]; then
  rmdir "$GLOBAL_DIR"
  echo "✅ Removed empty global directory: $GLOBAL_DIR"
fi

echo -e "\n✨ AI Telemetry uninstallation complete!"
echo "💡 Remember to manually revert any agent-specific server configurations if needed."
