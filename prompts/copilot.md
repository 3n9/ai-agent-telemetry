# GitHub Copilot CLI — AI Telemetry Instructions

Add this content to `.github/copilot-instructions.md` in your repository.

---

CRITICAL: Call `ai-log emit` on your FIRST tool-calling turn after every user message, in parallel with `report_intent` and any other initial tool calls. Do NOT start any work without logging first.

## What NOT to do (wrong)

❌ Reading files or running tools → then calling `ai-log emit` after  
❌ Skipping `ai-log emit` because the task "seems simple" or is just exploration  
❌ Calling `report_intent` without also calling `ai-log emit`

## What to do (correct)

✅ Call `ai-log emit` in parallel with `report_intent` and the first file read or tool use  
✅ Capture the returned `task_id` and use it as `parent_task_id` for subtasks or interruptions  
✅ Use `task_type: "subtask"` when beginning a distinct sub-step of a larger plan

## Emit command

```sh
ai-log emit '<json>'
ai-log emit --parent-task-id=<id> '<json>'
```

## State Management (Task Linking)

To link subtasks, you MUST capture the `task_id` from the parent's emit output. Use this pattern to ensure you don't lose the state:

```sh
# 1. Emit Parent and capture JSON response
PARENT_RESP=$(ai-log emit '{"schema_version":1, "agent_name":"copilot-cli", ... "task_type":"task"}')

# 2. Extract task_id into a variable (using python3 for reliability)
PARENT_ID=$(echo "$PARENT_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['task_id'])")
echo "Active Parent Task: $PARENT_ID"

# 3. Use variable for subtasks
ai-log emit --parent-task-id="$PARENT_ID" '{"schema_version":1, ... "task_type":"subtask"}'
```

## Required behavior

- ALWAYS log before or alongside the first tool call of any request — no exceptions
- NEVER skip logging, even for short, simple, or read-only tasks
- Capture the returned `task_id` from `ai-log emit` and reuse it as `parent_task_id` for subtasks and interruptions
- Prefer standard vocabulary for `work_type`, `language`, and `domain`
- Pass `agent_name: "copilot-cli"` and the active `model_name` in every payload

## Recommended values

**work_type:** `coding` · `debugging` · `research` · `analysis` · `writing` · `planning` · `creative` · `support` · `refactor`

**language:** `php` · `javascript` · `typescript` · `python` · `sql` · `html` · `css` · `shell` · `json` · `yaml` · `markdown` · `none`

**domain:** `frontend` · `backend` · `database` · `devops` · `documentation` · `wordpress` · `laravel` · `api` · `testing` · `fiction` · `horror` · `email` · `blog` · `marketing` · `none`

## Custom tags

Use `custom_tags` for concise task details that are not already covered by `work_type`, `language`, `domain`, or `secondary_work_type`.

- Prefer 0 to 3 tags; maximum 5
- Use short lowercase labels such as `auth`, `ui-polish`, `customer-reply`, `seo-draft`
- Favor themes, artifacts, audience, or intent over generic filler
- Do not repeat existing structured fields
- Do not include file paths, source code, prompts, or user messages

## Example — task + subtask

Call `ai-log emit` for the main task:

```json
{
  "agent_name": "copilot-cli",
  "model_name": "claude-3-5-sonnet",
  "work_type": "coding",
  "language": "typescript",
  "domain": "frontend",
  "complexity": "low",
  "confidence": 0.9,
  "estimated_time_min": 10,
  "task_type": "task",
  "custom_tags": ["component", "ui"]
}
```

Then use the returned ID for a subtask:

```json
{
  "agent_name": "copilot-cli",
  "model_name": "claude-3-5-sonnet",
  "parent_task_id": "01H...X",
  "work_type": "debugging",
  "language": "typescript",
  "domain": "frontend",
  "complexity": "low",
  "confidence": 0.85,
  "estimated_time_min": 5,
  "task_type": "subtask"
}
```

## Privacy

Never include file paths, source code, prompts, or user messages in any field.
