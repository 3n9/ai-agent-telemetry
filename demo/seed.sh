#!/usr/bin/env bash
# Seed demo/telemetry.db with realistic AI agent telemetry data.
# Usage: bash demo/seed.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
export AI_LOG_DB="$SCRIPT_DIR/telemetry.db"
AI_LOG="${AI_LOG:-ai-log}"

echo "🌱 Seeding demo database at $AI_LOG_DB..."
rm -f "$AI_LOG_DB"
"$AI_LOG" init

emit() { "$AI_LOG" emit "$1" > /dev/null; }

# ── claude-code / claude-sonnet-4.5 ─────────────────────────────────────────
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"coding","language":"typescript","domain":"frontend","complexity":"medium","confidence":0.85,"estimated_time_min":25,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"debugging","language":"typescript","domain":"frontend","complexity":"low","confidence":0.92,"estimated_time_min":10,"task_type":"subtask"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"coding","language":"python","domain":"backend","complexity":"high","confidence":0.7,"estimated_time_min":60,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"refactor","language":"python","domain":"backend","complexity":"medium","confidence":0.88,"estimated_time_min":30,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"debugging","language":"python","domain":"backend","complexity":"high","confidence":0.6,"estimated_time_min":45,"task_type":"interruption"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"analysis","language":"sql","domain":"database","complexity":"medium","confidence":0.9,"estimated_time_min":20,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"coding","language":"go","domain":"backend","complexity":"medium","confidence":0.82,"estimated_time_min":35,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"coding","language":"go","domain":"devops","complexity":"low","confidence":0.95,"estimated_time_min":15,"task_type":"subtask"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"writing","language":"markdown","domain":"documentation","complexity":"low","confidence":0.98,"estimated_time_min":10,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"research","language":"none","domain":"backend","complexity":"medium","confidence":0.75,"estimated_time_min":20,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"coding","language":"typescript","domain":"api","complexity":"high","confidence":0.78,"estimated_time_min":55,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"debugging","language":"typescript","domain":"api","complexity":"medium","confidence":0.83,"estimated_time_min":20,"task_type":"subtask"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"coding","language":"css","domain":"frontend","complexity":"low","confidence":0.96,"estimated_time_min":8,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-sonnet-4.5","work_type":"analysis","language":"python","domain":"database","complexity":"high","confidence":0.65,"estimated_time_min":50,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-opus-4.5","work_type":"planning","language":"none","domain":"backend","complexity":"high","confidence":0.72,"estimated_time_min":40,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-opus-4.5","work_type":"coding","language":"go","domain":"api","complexity":"high","confidence":0.8,"estimated_time_min":70,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"claude-code","model_name":"claude-opus-4.5","work_type":"refactor","language":"go","domain":"backend","complexity":"medium","confidence":0.87,"estimated_time_min":30,"task_type":"subtask"}'

# ── copilot-cli / gpt-4.1 ───────────────────────────────────────────────────
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"coding","language":"typescript","domain":"frontend","complexity":"medium","confidence":0.88,"estimated_time_min":22,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"debugging","language":"javascript","domain":"frontend","complexity":"medium","confidence":0.79,"estimated_time_min":18,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"coding","language":"python","domain":"backend","complexity":"low","confidence":0.93,"estimated_time_min":12,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"coding","language":"yaml","domain":"devops","complexity":"low","confidence":0.97,"estimated_time_min":8,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"analysis","language":"none","domain":"devops","complexity":"medium","confidence":0.84,"estimated_time_min":25,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"coding","language":"typescript","domain":"api","complexity":"medium","confidence":0.86,"estimated_time_min":28,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"debugging","language":"typescript","domain":"api","complexity":"low","confidence":0.91,"estimated_time_min":10,"task_type":"subtask"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"writing","language":"markdown","domain":"documentation","complexity":"low","confidence":0.99,"estimated_time_min":5,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"refactor","language":"typescript","domain":"frontend","complexity":"medium","confidence":0.85,"estimated_time_min":20,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-4.1","work_type":"coding","language":"sql","domain":"database","complexity":"medium","confidence":0.88,"estimated_time_min":15,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-5-mini","work_type":"coding","language":"python","domain":"backend","complexity":"low","confidence":0.94,"estimated_time_min":10,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-5-mini","work_type":"debugging","language":"python","domain":"backend","complexity":"medium","confidence":0.77,"estimated_time_min":22,"task_type":"interruption"}'
emit '{"schema_version":1,"agent_name":"copilot-cli","model_name":"gpt-5-mini","work_type":"analysis","language":"none","domain":"backend","complexity":"low","confidence":0.9,"estimated_time_min":12,"task_type":"task"}'

# ── gemini-cli / gemini-2.0-flash ───────────────────────────────────────────
emit '{"schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.0-flash","work_type":"coding","language":"python","domain":"backend","complexity":"medium","confidence":0.82,"estimated_time_min":28,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.0-flash","work_type":"analysis","language":"python","domain":"database","complexity":"medium","confidence":0.87,"estimated_time_min":18,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.0-flash","work_type":"coding","language":"go","domain":"backend","complexity":"low","confidence":0.93,"estimated_time_min":14,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.0-flash","work_type":"research","language":"none","domain":"devops","complexity":"medium","confidence":0.76,"estimated_time_min":30,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.0-flash","work_type":"debugging","language":"go","domain":"backend","complexity":"medium","confidence":0.81,"estimated_time_min":20,"task_type":"subtask"}'
emit '{"schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.0-pro","work_type":"coding","language":"typescript","domain":"frontend","complexity":"high","confidence":0.73,"estimated_time_min":50,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.0-pro","work_type":"planning","language":"none","domain":"api","complexity":"high","confidence":0.68,"estimated_time_min":45,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.0-pro","work_type":"coding","language":"python","domain":"database","complexity":"high","confidence":0.71,"estimated_time_min":60,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"gemini-cli","model_name":"gemini-2.0-pro","work_type":"refactor","language":"typescript","domain":"frontend","complexity":"medium","confidence":0.85,"estimated_time_min":25,"task_type":"subtask"}'

# ── codex-cli / gpt-5.3-codex ───────────────────────────────────────────────
emit '{"schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5.3-codex","work_type":"coding","language":"javascript","domain":"frontend","complexity":"medium","confidence":0.86,"estimated_time_min":20,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5.3-codex","work_type":"coding","language":"python","domain":"backend","complexity":"medium","confidence":0.84,"estimated_time_min":25,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5.3-codex","work_type":"debugging","language":"javascript","domain":"frontend","complexity":"low","confidence":0.93,"estimated_time_min":8,"task_type":"subtask"}'
emit '{"schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5.3-codex","work_type":"coding","language":"shell","domain":"devops","complexity":"low","confidence":0.97,"estimated_time_min":6,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5.3-codex","work_type":"refactor","language":"javascript","domain":"frontend","complexity":"medium","confidence":0.83,"estimated_time_min":18,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5.3-codex","work_type":"analysis","language":"sql","domain":"database","complexity":"medium","confidence":0.88,"estimated_time_min":15,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5.3-codex","work_type":"coding","language":"typescript","domain":"api","complexity":"high","confidence":0.74,"estimated_time_min":55,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5.3-codex","work_type":"writing","language":"markdown","domain":"documentation","complexity":"low","confidence":0.98,"estimated_time_min":7,"task_type":"task"}'
emit '{"schema_version":1,"agent_name":"codex-cli","model_name":"gpt-5.3-codex","work_type":"support","language":"none","domain":"none","complexity":"low","confidence":0.95,"estimated_time_min":5,"task_type":"task"}'

echo "✅ Demo database seeded."
echo "   DB: $AI_LOG_DB"
echo ""
echo "Preview:"
AI_LOG_DB="$AI_LOG_DB" "$AI_LOG" init 2>/dev/null || true
AI_LOG_DB="$AI_LOG_DB" ai-log-report summary 2>/dev/null || true
