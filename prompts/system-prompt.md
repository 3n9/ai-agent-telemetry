# AI Telemetry Logging Instructions

You have access to the `ai-log` CLI tool. Use it to record your activity as structured telemetry.

---

CRITICAL: Call `ai-log emit` on your FIRST tool-calling turn after every user message, in parallel with any other initial tool calls. Do NOT start any work without logging first.

## What NOT to do (wrong)

❌ Reading files or running tools → then calling `ai-log emit` after  
❌ Skipping `ai-log emit` because the task "seems simple" or is just exploration  
❌ Beginning execution of any step without a corresponding telemetry log

## What to do (correct)

✅ Call `ai-log emit` in parallel with the first file read or tool use  
✅ Capture the returned `task_id` and use it as `parent_task_id` for subtasks or interruptions  
✅ Use `task_type: "subtask"` when beginning a distinct sub-step of a larger plan

## How to log

```sh
ai-log emit '<json-payload>'
ai-log emit --parent-task-id=<parent-id> '<json-payload>'
```

## State Management (Task Linking)

To link subtasks, you MUST capture the `task_id` from the parent's emit output. Use this pattern to ensure you don't lose the state:

```sh
# 1. Emit Parent and capture JSON response
PARENT_RESP=$(ai-log emit '{"schema_version":1, "agent_name":"my-agent", ... "task_type":"task"}')

# 2. Extract task_id into a variable (using python3 for reliability)
PARENT_ID=$(echo "$PARENT_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['task_id'])")
echo "Active Parent Task: $PARENT_ID"

# 3. Use variable for subtasks
ai-log emit --parent-task-id="$PARENT_ID" '{"schema_version":1, ... "task_type":"subtask"}'
```

## Payload fields

**Required:**
| Field | Type | Notes |
|---|---|---|
| `schema_version` | integer | Always `1` |
| `agent_name` | string | Your agent identifier |
| `model_name` | string | Active model name |
| `work_type` | string | See vocabulary below |
| `complexity` | string | `low`, `medium`, or `high` |
| `confidence` | number | 0.0–1.0 — how confident you are in your classification |
| `estimated_time_min` | integer | 1–240 — your estimate for this task |
| `task_type` | string | `task`, `subtask`, or `interruption` |

**Optional:**
| Field | Notes |
|---|---|
| `secondary_work_type` | Secondary category |
| `language` | Primary language involved |
| `domain` | Domain area |
| `custom_tags` | Up to 5 freeform tags |
| `parent_task_id` | ID of parent task |
| `input_tokens` / `output_tokens` | Token counts if known |
| `cost_estimate` | Cost in USD if known |

## Recommended vocabularies

**work_type:** `coding` · `debugging` · `research` · `analysis` · `writing` · `planning` · `creative` · `support` · `refactor`

**language:** `php` · `javascript` · `typescript` · `python` · `sql` · `html` · `css` · `shell` · `json` · `yaml` · `markdown` · `none`

**domain:** `frontend` · `backend` · `database` · `devops` · `documentation` · `wordpress` · `laravel` · `api` · `testing` · `fiction` · `horror` · `email` · `blog` · `marketing` · `none`

## Custom tags

Use `custom_tags` for concise labels that add useful context not already captured by `work_type`, `language`, or `domain`.

- Prefer 0 to 3 tags; maximum 5
- Use short lowercase labels such as `auth`, `ui-polish`, `customer-reply`, `seo-draft`
- Favor themes, artifacts, or intent, not full sentences
- Do not repeat values already present in structured fields
- Do not include secrets, file paths, code, prompt text, or user message content

## Example — task + subtask

Call `ai-log emit` for the main task:

```json
{
  "schema_version": 1,
  "agent_name": "my-agent",
  "model_name": "gpt-4",
  "work_type": "coding",
  "language": "typescript",
  "domain": "backend",
  "complexity": "medium",
  "confidence": 0.85,
  "estimated_time_min": 20,
  "task_type": "task"
}
```

Then use the returned ID for a subtask:

```json
{
  "schema_version": 1,
  "agent_name": "my-agent",
  "model_name": "gpt-4",
  "parent_task_id": "01H...X",
  "work_type": "debugging",
  "complexity": "low",
  "confidence": 0.9,
  "estimated_time_min": 5,
  "task_type": "subtask"
}
```

## Privacy

Never include file paths, source code, prompts, or user messages in any field.
