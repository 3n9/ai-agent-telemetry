#!/usr/bin/env bash
# Seed demo/telemetry.db with realistic AI agent telemetry data (~400 records,
# spread across the last 30 days). Uses Python for fast direct DB insertion.
# Usage: bash demo/seed.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
export AI_LOG_DB="$SCRIPT_DIR/telemetry.db"
AI_LOG="${AI_LOG:-ai-log}"

echo "🌱 Seeding demo database at $AI_LOG_DB..."
rm -f "$AI_LOG_DB"
"$AI_LOG" init

python3 << 'PYEOF'
import sqlite3, json, random, os
from datetime import datetime, timezone, timedelta

DB = os.environ["AI_LOG_DB"]
con = sqlite3.connect(DB)

# ── vocabulary ────────────────────────────────────────────────────────────────
AGENTS = {
    "claude-code":  ["claude-sonnet-4.5", "claude-sonnet-4.5", "claude-opus-4.5"],
    "copilot-cli":  ["claude-sonnet-4.6", "claude-sonnet-4.6", "gpt-4.1", "gpt-5-mini"],
    "gemini-cli":   ["gemini-2.0-flash", "gemini-2.0-flash", "gemini-2.0-pro"],
    "codex-cli":    ["gpt-5.3-codex", "gpt-5.3-codex", "gpt-5.1-codex"],
}
RECOMMENDED_WORK_TYPES = {"coding","debugging","refactor","analysis","planning","research","writing","support"}
RECOMMENDED_LANGS      = {"typescript","python","go","javascript","sql","yaml","shell","markdown","css","none"}
RECOMMENDED_DOMAINS    = {"frontend","backend","api","database","devops","documentation","testing","none"}

# explicit order so weights align correctly
WORK_TYPES   = ["coding","debugging","refactor","analysis","planning","research","writing","support"]
WT_WEIGHTS   = [35,      18,         10,        10,        6,         6,         10,       5]
LANGUAGES    = ["typescript","python","go","javascript","sql","yaml","shell","markdown","css","none"]
DOMAINS      = ["frontend","backend","api","database","devops","documentation","testing","none"]
COMPLEXITIES = ["low", "low", "medium", "medium", "medium", "high"]

CUSTOM_TAGS_POOL = [
    ["auth", "jwt"],
    ["ui-polish"],
    ["performance"],
    ["migration", "schema"],
    ["ci-cd"],
    ["security"],
    ["onboarding"],
    ["seo"],
    ["caching"],
    ["rate-limiting"],
    ["logging"],
    ["monitoring"],
    ["feature-flag"],
    ["a11y"],
    [],
    [],
    [],  # no tags — most common
]

NOW = datetime.now(timezone.utc)

def rand_date(days_back=30):
    offset = random.randint(0, days_back * 86400)
    return (NOW - timedelta(seconds=offset)).strftime("%Y-%m-%dT%H:%M:%SZ")

def tag_source(val, vocab):
    return "recommended" if val in vocab else "custom"

def ulid_fake(created_at: str) -> str:
    """Fake ULID: timestamp prefix + random suffix (good enough for demo)."""
    import time as _time
    ts = int(datetime.strptime(created_at, "%Y-%m-%dT%H:%M:%SZ")
             .replace(tzinfo=timezone.utc).timestamp() * 1000)
    chars = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
    t = ""
    for _ in range(10):
        t = chars[ts & 31] + t
        ts >>= 5
    r = "".join(random.choices(chars, k=16))
    return t + r

def make_record(agent, task_type, parent_id=None):
    model      = random.choice(AGENTS[agent])
    work_type  = random.choices(WORK_TYPES, weights=WT_WEIGHTS)[0]
    language   = random.choice(LANGUAGES)
    domain     = random.choice(DOMAINS)
    complexity = random.choice(COMPLEXITIES)
    confidence = round(random.uniform(0.55, 0.99), 2)
    minutes    = random.choice([5,8,10,12,15,18,20,22,25,28,30,35,40,45,50,55,60,70,80])
    created_at = rand_date()
    tid        = ulid_fake(created_at)
    tags       = random.choice(CUSTOM_TAGS_POOL)

    payload = {
        "schema_version": 1,
        "agent_name":     agent,
        "model_name":     model,
        "work_type":      work_type,
        "language":       language,
        "domain":         domain,
        "complexity":     complexity,
        "confidence":     confidence,
        "estimated_time_min": minutes,
        "task_type":      task_type,
    }
    if tags:
        payload["custom_tags"] = tags
    if parent_id:
        payload["parent_task_id"] = parent_id

    row = (
        tid, created_at, 1,
        agent, model,
        work_type, tag_source(work_type, RECOMMENDED_WORK_TYPES),
        None, None,
        language, tag_source(language, RECOMMENDED_LANGS),
        domain, tag_source(domain, RECOMMENDED_DOMAINS),
        complexity, confidence, minutes,
        task_type, parent_id,
        None, None, None,
        json.dumps(payload),
    )
    return tid, row, tags, created_at

INSERT = """
INSERT INTO tasks (
  id, created_at, schema_version,
  agent_name, model_name,
  work_type, work_type_tag_source,
  secondary_work_type, secondary_work_type_tag_source,
  language, language_tag_source,
  domain, domain_tag_source,
  complexity, confidence, estimated_time_min,
  task_type, parent_task_id,
  input_tokens, output_tokens, cost_estimate,
  raw_payload_json
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
"""

TAG_INSERT = "INSERT INTO task_tags (task_id, tag_value, tag_source) VALUES (?,?,?)"

agents = list(AGENTS.keys())

# ── emit tasks (150) ─────────────────────────────────────────────────────────
task_ids = []
for _ in range(150):
    agent = random.choice(agents)
    tid, row, tags, cat = make_record(agent, "task")
    con.execute(INSERT, row)
    for tag in tags:
        src = "recommended" if tag in {"auth","ci-cd","security","logging","monitoring","performance"} else "custom"
        con.execute(TAG_INSERT, (tid, tag, src))
    task_ids.append(tid)

# ── emit subtasks (230) ───────────────────────────────────────────────────────
for _ in range(230):
    agent  = random.choice(agents)
    parent = random.choice(task_ids)
    tid, row, tags, cat = make_record(agent, "subtask", parent_id=parent)
    con.execute(INSERT, row)
    for tag in tags:
        src = "recommended" if tag in {"auth","ci-cd","security","logging","monitoring","performance"} else "custom"
        con.execute(TAG_INSERT, (tid, tag, src))

# ── emit interruptions (25) ───────────────────────────────────────────────────
for _ in range(25):
    agent  = random.choice(agents)
    parent = random.choice(task_ids) if random.random() > 0.4 else None
    tid, row, tags, cat = make_record(agent, "interruption", parent_id=parent)
    con.execute(INSERT, row)

con.commit()
con.close()
print(f"  inserted 405 records (150 tasks · 230 subtasks · 25 interruptions)")
PYEOF

echo ""
echo "✅ Demo database seeded."
echo "   DB: $AI_LOG_DB"
echo ""
echo "Preview:"
AI_LOG_DB="$AI_LOG_DB" ai-log-report summary 2>/dev/null || true
