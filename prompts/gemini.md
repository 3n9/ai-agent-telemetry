# Gemini CLI — AI Telemetry Instructions

Add this file as `GEMINI.md` in your project root (or append to an existing one).

---

CRITICAL: Call `ai-log emit` on your FIRST tool-calling turn after every user message, in parallel with any other initial tool calls. Do NOT start any work without logging first.

Note: A `BeforeAgent` hook also injects this policy automatically — but you must still call `ai-log emit` yourself on every request to ensure you have a task context.

## What NOT to do (wrong)

❌ Reading files or running tools → then calling `ai-log emit` after  
❌ Skipping `ai-log emit` because the task "seems simple" or is just exploration  
❌ Omitting `agent_name` and `model_name` from the JSON payload

## What to do (correct)

✅ Call `ai-log emit` in parallel with the first file read or tool use  
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
PARENT_RESP=$(ai-log emit '{"schema_version":1, "agent_name":"gemini-cli", ... "task_type":"task"}')

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
- Pass `agent_name: "gemini-cli"` and the active `model_name` in every payload

## Recommended values

**work_type:** `coding` · `debugging` · `research` · `analysis` · `writing` · `planning` · `creative` · `support` · `refactor`

**language:** `php` · `javascript` · `typescript` · `python` · `sql` · `html` · `css` · `shell` · `json` · `yaml` · `markdown` · `none`

**domain:** `frontend` · `backend` · `database` · `devops` · `documentation` · `wordpress` · `laravel` · `api` · `testing` · `fiction` · `horror` · `email` · `blog` · `marketing` · `none`

## Custom tags

Use `custom_tags` for useful detail that does not fit cleanly into `work_type`, `language`, `domain`, or `secondary_work_type`.

- Prefer 0 to 3 tags; maximum 5
- Use short lowercase labels such as `auth`, `ui-polish`, `customer-reply`, `seo-draft`
- Favor themes, artifacts, audience, or intent over vague labels
- Do not repeat values already present in structured fields
- Do not include file paths, source code, prompts, or user messages

## Example — research task

```sh
RESPONSE=$(ai-log emit '{
  "schema_version": 1,
  "agent_name": "gemini-cli",
  "model_name": "gemini-2.0-flash-exp",
  "work_type": "research",
  "secondary_work_type": "analysis",
  "language": "python",
  "domain": "backend",
  "complexity": "high",
  "confidence": 0.75,
  "estimated_time_min": 30,
  "task_type": "task",
  "custom_tags": ["architecture", "evaluation"]
}')
```

## Example — task + subtask

Call `ai-log emit` for the main task:

```json
{
  "agent_name": "gemini-cli",
  "model_name": "gemini-2.0-flash-exp",
  "work_type": "coding",
  "language": "typescript",
  "domain": "backend",
  "complexity": "medium",
  "confidence": 0.8,
  "estimated_time_min": 25,
  "task_type": "task",
  "custom_tags": ["auth"]
}
```

Then use the returned ID for a subtask:

```json
{
  "agent_name": "gemini-cli",
  "model_name": "gemini-2.0-flash-exp",
  "parent_task_id": "01H...X",
  "work_type": "debugging",
  "language": "typescript",
  "domain": "backend",
  "complexity": "low",
  "confidence": 0.9,
  "estimated_time_min": 8,
  "task_type": "subtask"
}
```

## Privacy

Never include file paths, source code, prompts, or user messages in any field.
