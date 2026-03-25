#!/usr/bin/env bash
# scripts/seed-test-data.sh
# Generates a realistic spread of telemetry records using the ai-log CLI.
# Usage: ./scripts/seed-test-data.sh [path/to/ai-log]

set -euo pipefail

AI_LOG="${1:-ai-log}"

if ! command -v "$AI_LOG" &>/dev/null && [[ ! -x "$AI_LOG" ]]; then
  echo "ERROR: ai-log not found. Pass path as first argument or put it on \$PATH." >&2
  exit 1
fi

emit() { "$AI_LOG" emit "$1"; }

task_id() {
  local json="$1"
  echo "$json" | grep '"task_id"' | sed 's/.*"task_id": *"\([^"]*\)".*/\1/'
}

echo "==> Initialising database..."
"$AI_LOG" init

echo ""
echo "==> Seeding tasks..."

# ── Claude Code — coding ──────────────────────────────────────────────────────
T1=$(emit '{
  "schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4-5",
  "work_type":"coding","secondary_work_type":"refactor",
  "language":"typescript","domain":"backend",
  "complexity":"high","confidence":0.82,"estimated_time_min":45,
  "task_type":"task","custom_tags":["api","rest"]
}')
T1_ID=$(task_id "$T1")
echo "  task: $T1_ID  (coding/typescript/backend)"

emit "{
  \"schema_version\":1,\"agent_name\":\"claude-code\",\"model_name\":\"claude-sonnet-4-5\",
  \"work_type\":\"debugging\",\"language\":\"typescript\",\"domain\":\"backend\",
  \"complexity\":\"medium\",\"confidence\":0.88,\"estimated_time_min\":15,
  \"task_type\":\"subtask\",\"parent_task_id\":\"$T1_ID\"
}" > /dev/null
echo "  subtask linked to $T1_ID  (debugging)"

# ── Claude Code — interruption ────────────────────────────────────────────────
emit "{
  \"schema_version\":1,\"agent_name\":\"claude-code\",\"model_name\":\"claude-sonnet-4-5\",
  \"work_type\":\"analysis\",\"language\":\"typescript\",\"domain\":\"backend\",
  \"complexity\":\"medium\",\"confidence\":0.41,\"estimated_time_min\":10,
  \"task_type\":\"interruption\",\"parent_task_id\":\"$T1_ID\",
  \"custom_tags\":[\"blocked\",\"missing-context\"]
}" > /dev/null
echo "  interruption linked to $T1_ID  (blocked)"

# ── Copilot — frontend tasks ──────────────────────────────────────────────────
T2=$(emit '{
  "schema_version":1,"agent_name":"copilot-cli","model_name":"claude-sonnet-4-6",
  "work_type":"coding","secondary_work_type":"analysis",
  "language":"javascript","domain":"frontend",
  "complexity":"medium","confidence":0.9,"estimated_time_min":20,
  "task_type":"task","custom_tags":["ui-polish","components"]
}')
T2_ID=$(task_id "$T2")
echo "  task: $T2_ID  (coding/javascript/frontend)"

emit '{
  "schema_version":1,"agent_name":"copilot-cli","model_name":"claude-sonnet-4-6",
  "work_type":"debugging","language":"css","domain":"frontend",
  "complexity":"low","confidence":0.95,"estimated_time_min":8,
  "task_type":"task"
}' > /dev/null
echo "  task: (debugging/css/frontend)"

emit '{
  "schema_version":1,"agent_name":"copilot-cli","model_name":"claude-sonnet-4-6",
  "work_type":"refactor","language":"javascript","domain":"frontend",
  "complexity":"medium","confidence":0.78,"estimated_time_min":30,
  "task_type":"task","custom_tags":["cleanup"]
}' > /dev/null
echo "  task: (refactor/javascript/frontend)"

# ── Gemini — research & creative ─────────────────────────────────────────────
T3=$(emit '{
  "schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.5-pro",
  "work_type":"research","secondary_work_type":"analysis",
  "language":"python","domain":"backend",
  "complexity":"high","confidence":0.7,"estimated_time_min":60,
  "task_type":"task","custom_tags":["architecture","evaluation"]
}')
T3_ID=$(task_id "$T3")
echo "  task: $T3_ID  (research/python/backend)"

emit "{
  \"schema_version\":1,\"agent_name\":\"gemini-cli\",\"model_name\":\"gemini-2.5-pro\",
  \"work_type\":\"writing\",\"language\":\"markdown\",\"domain\":\"documentation\",
  \"complexity\":\"low\",\"confidence\":0.91,\"estimated_time_min\":12,
  \"task_type\":\"subtask\",\"parent_task_id\":\"$T3_ID\"
}" > /dev/null
echo "  subtask linked to $T3_ID  (writing/docs)"

emit '{
  "schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.5-pro",
  "work_type":"creative","secondary_work_type":"writing",
  "language":"markdown","domain":"fiction",
  "complexity":"medium","confidence":0.77,"estimated_time_min":14,
  "task_type":"task","custom_tags":["scene-idea","character-voice"]
}' > /dev/null
echo "  task: (creative/markdown/fiction)"

# ── Codex — mixed workload ────────────────────────────────────────────────────
T4=$(emit '{
  "schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5",
  "work_type":"coding","language":"python","domain":"api",
  "complexity":"high","confidence":0.85,"estimated_time_min":50,
  "task_type":"task","custom_tags":["auth","jwt"]
}')
T4_ID=$(task_id "$T4")
echo "  task: $T4_ID  (coding/python/api)"

emit "{
  \"schema_version\":1,\"agent_name\":\"codex-cli\",\"model_name\":\"gpt-5\",
  \"work_type\":\"debugging\",\"language\":\"python\",\"domain\":\"api\",
  \"complexity\":\"medium\",\"confidence\":0.8,\"estimated_time_min\":18,
  \"task_type\":\"subtask\",\"parent_task_id\":\"$T4_ID\"
}" > /dev/null
echo "  subtask linked to $T4_ID  (debugging)"

emit "{
  \"schema_version\":1,\"agent_name\":\"codex-cli\",\"model_name\":\"gpt-5\",
  \"work_type\":\"analysis\",\"language\":\"python\",\"domain\":\"api\",
  \"complexity\":\"low\",\"confidence\":0.35,\"estimated_time_min\":7,
  \"task_type\":\"interruption\",\"parent_task_id\":\"$T4_ID\",
  \"custom_tags\":[\"dependency-failure\"]
}" > /dev/null
echo "  interruption linked to $T4_ID  (dependency-failure)"

emit '{
  "schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5",
  "work_type":"planning","language":"markdown","domain":"devops",
  "complexity":"medium","confidence":0.88,"estimated_time_min":22,
  "task_type":"task","custom_tags":["ci","deployment"]
}' > /dev/null
echo "  task: (planning/devops)"

emit '{
  "schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5",
  "work_type":"coding","language":"yaml","domain":"devops",
  "complexity":"low","confidence":0.93,"estimated_time_min":10,
  "task_type":"task"
}' > /dev/null
echo "  task: (coding/yaml/devops)"

# ── Claude Code — writing & support ──────────────────────────────────────────
emit '{
  "schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4-5",
  "work_type":"writing","secondary_work_type":"planning",
  "language":"markdown","domain":"blog",
  "complexity":"medium","confidence":0.83,"estimated_time_min":16,
  "task_type":"task","custom_tags":["outline","seo-draft"]
}' > /dev/null
echo "  task: (writing/markdown/blog)"

emit '{
  "schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4-5",
  "work_type":"support","language":"none","domain":"documentation",
  "complexity":"low","confidence":0.96,"estimated_time_min":5,
  "task_type":"task"
}' > /dev/null
echo "  task: (support/documentation)"

emit '{
  "schema_version":1,"agent_name":"claude-code","model_name":"claude-opus-4-5",
  "work_type":"analysis","secondary_work_type":"research",
  "language":"sql","domain":"database",
  "complexity":"high","confidence":0.72,"estimated_time_min":40,
  "task_type":"task","custom_tags":["performance","query-optimisation"]
}' > /dev/null
echo "  task: (analysis/sql/database)"

# ── Standalone interruption (no parent) ──────────────────────────────────────
emit '{
  "schema_version":1,"agent_name":"copilot-cli","model_name":"claude-sonnet-4-6",
  "work_type":"coding","language":"typescript","domain":"frontend",
  "complexity":"medium","confidence":0.38,"estimated_time_min":9,
  "task_type":"interruption","custom_tags":["tool-malfunction"]
}' > /dev/null
echo "  interruption: standalone (tool-malfunction)"

echo ""
echo "==> Done. Run 'ai-log-report summary' to view results."
