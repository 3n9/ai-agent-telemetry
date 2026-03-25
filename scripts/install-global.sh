#!/usr/bin/env bash
# Inject AI Telemetry agent prompts into global agent config files.
#
# Usage (from a repo clone):
#   make install-global
#
# Usage (standalone, no clone required):
#   curl -fsSL https://raw.githubusercontent.com/3n9/ai-agent-telemetry/main/scripts/install-global.sh | sh
set -e

REPO="3n9/ai-agent-telemetry"
RAW_BASE="https://raw.githubusercontent.com/$REPO/main"

# 1. SETUP PATHS
# When invoked via curl, $0 is "sh" or "-" so dirname gives ".".
# Detect whether we're running from a real repo clone by checking for prompts/.
_SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$_SCRIPT_DIR/.." && pwd)"
GLOBAL_DIR="$HOME/.ai-telemetry"
PROMPTS_DEST="$GLOBAL_DIR/prompts"
BIN_DEST="$GLOBAL_DIR/bin"
BLOCK_NAME="AI TELEMETRY SYSTEM (AUTO-GENERATED)"
BEGIN_MARKER="### BEGIN $BLOCK_NAME"
END_MARKER="### END $BLOCK_NAME"

echo "🚀 Installing AI Telemetry Globally..."

# 2. SYNC PROMPTS AND SCRIPTS TO HOME
# If running from a local clone, copy from disk. Otherwise fetch from GitHub.
mkdir -p "$PROMPTS_DEST" "$BIN_DEST"

_fetch_file() {
    local name="$1"
    local dest="$2"
    local local_src="$3"
    if [ -f "$local_src" ]; then
        cp "$local_src" "$dest"
    else
        echo "   Fetching $name from GitHub..."
        curl -fsSL "$RAW_BASE/$name" -o "$dest"
    fi
}

for md in claude-code.md copilot.md codex.md gemini.md system-prompt.md; do
    _fetch_file "prompts/$md" "$PROMPTS_DEST/$md" "$PROJECT_ROOT/prompts/$md"
done
_fetch_file "scripts/gemini-telemetry-hook.py" "$BIN_DEST/gemini-telemetry-hook.py" \
    "$PROJECT_ROOT/scripts/gemini-telemetry-hook.py"
chmod +x "$BIN_DEST/gemini-telemetry-hook.py"
echo "✅ Prompts and hook script synced to $GLOBAL_DIR"

# 3. BUILD AND INSTALL BINARIES (only when running from a clone)
if [ -f "$PROJECT_ROOT/go.mod" ]; then
    BIN_INSTALL="${INSTALL_DIR:-$HOME/.local/bin}"
    mkdir -p "$BIN_INSTALL"
    echo "🔨 Building binaries from source..."
    for cmd in ai-log ai-log-report; do
        go build -o "$BIN_INSTALL/$cmd" "$PROJECT_ROOT/cmd/$cmd"
        echo "   ✅ $cmd → $BIN_INSTALL/$cmd"
    done
fi

# 3. HELPER: SAFE INJECT (sync — replaces existing block if present)
safe_inject() {
    local target="$1"
    local source="$2"
    local name="$3"

    mkdir -p "$(dirname "$target")"
    echo "🔧 Syncing $name instructions in $target..."
    python3 - "$target" "$source" "$BEGIN_MARKER" "$END_MARKER" <<'PY'
import pathlib
import sys

target_path = pathlib.Path(sys.argv[1]).expanduser()
source_path = pathlib.Path(sys.argv[2]).expanduser()
begin_marker = sys.argv[3]
end_marker = sys.argv[4]

existing = target_path.read_text() if target_path.exists() else ""
source = source_path.read_text().rstrip("\n")
block = f"{begin_marker}\n{source}\n{end_marker}"

start = existing.find(begin_marker)
end = existing.find(end_marker)
if start != -1 and end != -1 and end >= start:
    end += len(end_marker)
    updated = existing[:start].rstrip("\n") + "\n\n" + block + existing[end:]
else:
    if existing and not existing.endswith("\n"):
        existing += "\n"
    separator = "\n" if existing.strip() else ""
    updated = existing + separator + block + "\n"

target_path.write_text(updated)
PY
    echo "✅ $name instructions synced."
}

# 4. AGENT: CLAUDE CODE
safe_inject "$HOME/.claude/CLAUDE.md" "$PROMPTS_DEST/claude-code.md" "Claude Code"

# 5. AGENT: COPILOT CLI
safe_inject "$HOME/.copilot/copilot-instructions.md" "$PROMPTS_DEST/copilot.md" "Copilot CLI"

# 6. AGENT: CODEX CLI
safe_inject "$HOME/.codex/AGENTS.md" "$PROMPTS_DEST/codex.md" "Codex"

# 7. AGENT: AIDER
AIDER_CONV="$HOME/.aider.conventions.md"
safe_inject "$AIDER_CONV" "$PROMPTS_DEST/system-prompt.md" "Aider Conventions"

AIDER_CONF="$HOME/.aider.conf.yml"
if [ ! -f "$AIDER_CONF" ] || ! grep -q ".aider.conventions.md" "$AIDER_CONF"; then
    echo "🔗 Linking Aider conventions in $AIDER_CONF..."
    mkdir -p "$(dirname "$AIDER_CONF")"
    [ ! -f "$AIDER_CONF" ] && echo "---" > "$AIDER_CONF"
    echo -e "read: [ \"$AIDER_CONV\" ]" >> "$AIDER_CONF"
fi

# 8. AGENT: GEMINI CLI (BeforeAgent Hook)
GEMINI_SETTINGS="$HOME/.gemini/settings.json"
mkdir -p "$(dirname "$GEMINI_SETTINGS")"
if [ ! -f "$GEMINI_SETTINGS" ]; then
    echo '{"hooks": {"BeforeAgent": []}}' > "$GEMINI_SETTINGS"
fi

# Patch Gemini settings.json using python for safety
python3 -c "
import json, os
path = os.path.expanduser('~/.gemini/settings.json')
hook_path = os.path.expanduser('~/.ai-telemetry/bin/gemini-telemetry-hook.py')
with open(path, 'r+') as f:
    data = json.load(f)
    hooks = data.setdefault('hooks', {}).setdefault('BeforeAgent', [])
    if not any(h.get('name') == 'ai-telemetry-policy' for h in hooks):
        hooks.append({
            'name': 'ai-telemetry-policy',
            'matcher': '*',
            'hooks': [{
                'name': 'telemetry-reminder',
                'type': 'command',
                'command': f'python3 {hook_path}'
            }]
        })
    f.seek(0)
    json.dump(data, f, indent=2)
    f.truncate()
"
echo "✅ Gemini CLI BeforeAgent hook configured."

echo -e "\n✨ Global installation complete!"
echo "💡 To use Aider with global conventions, ensure ~/.aider.conf.yml is valid."
