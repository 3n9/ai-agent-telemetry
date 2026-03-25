# ai-log Specification

## Overview

`ai-log` is the telemetry ingestion CLI tool.

Responsibilities:

- validate payloads
- normalize tags
- generate task IDs
- derive tag sources
- store telemetry in SQLite
- return structured JSON responses

---

## Commands

### Initialize database

ai-log init

Creates the telemetry database and tables.

---

### Emit telemetry record

ai-log emit '<json-payload>'

Stores a telemetry record.

If `task_id` is missing the tool generates one.

---

### Emit with parent task

ai-log emit --parent-task-id=<id> '<json>'

Automatically attaches a parent task ID.

---

### Emit with explicit task id

ai-log emit --task-id=<id> '<json>'

Allows callers to provide their own ID.

---

### Validate payload

ai-log validate '<json>'

Validates payload without writing to the database.

---

## Output Format

All responses are JSON.

Example success response:

```json
{
  "ok": true,
  "task_id": "01HTK3A4ZP6XQW9V5M7N2D4B8R",
  "parent_task_id": null,
  "task_type": "task",
  "schema_version": 1,
  "warnings": []
}
```

Example error:

```json
{
  "ok": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "confidence must be between 0 and 1"
  }
}
```

---

## Task ID generation

IDs are generated using **ULID** when missing.

Advantages:

- globally unique
- time sortable
- compact

---

## Tag Normalization

Rules applied during ingestion:

- lowercase
- trim whitespace
- replace spaces with "-"
- remove invalid characters
- deduplicate

Example:

`"Bug Fixing" → "bug-fixing"`

---

## Tag Source Derivation

Tags are classified as:

- recommended
- custom

Classification occurs during ingestion based on the starter vocabulary.

---

## Example Payloads

### Coding task

```json
{
  "schema_version": 1,
  "agent_name": "codex-cli",
  "model_name": "gpt-5",
  "work_type": "coding",
  "secondary_work_type": "analysis",
  "language": "javascript",
  "domain": "frontend",
  "custom_tags": ["ui-polish"],
  "complexity": "medium",
  "confidence": 0.82,
  "estimated_time_min": 18,
  "task_type": "task",
  "parent_task_id": null,
  "input_tokens": null,
  "output_tokens": null,
  "cost_estimate": null
}
```

### Writing task

```json
{
  "schema_version": 1,
  "agent_name": "claude-code",
  "model_name": "claude-sonnet",
  "work_type": "writing",
  "secondary_work_type": "planning",
  "language": "markdown",
  "domain": "blog",
  "custom_tags": ["outline", "seo-draft"],
  "complexity": "medium",
  "confidence": 0.83,
  "estimated_time_min": 16,
  "task_type": "task",
  "parent_task_id": null,
  "input_tokens": null,
  "output_tokens": null,
  "cost_estimate": null
}
```

### Creative task

```json
{
  "schema_version": 1,
  "agent_name": "gemini-cli",
  "model_name": "gemini-2.5-pro",
  "work_type": "creative",
  "secondary_work_type": "writing",
  "language": "markdown",
  "domain": "fiction",
  "custom_tags": ["scene-idea", "character-voice"],
  "complexity": "medium",
  "confidence": 0.77,
  "estimated_time_min": 14,
  "task_type": "task",
  "parent_task_id": null,
  "input_tokens": null,
  "output_tokens": null,
  "cost_estimate": null
}
```

### Interruption

```json
{
  "schema_version": 1,
  "agent_name": "claude-code",
  "model_name": "claude-sonnet",
  "work_type": "analysis",
  "secondary_work_type": null,
  "language": "python",
  "domain": "api",
  "custom_tags": ["blocked", "missing-context"],
  "complexity": "medium",
  "confidence": 0.41,
  "estimated_time_min": 9,
  "task_type": "interruption",
  "parent_task_id": "01HTK3A4ZP6XQW9V5M7N2D4B8R",
  "input_tokens": null,
  "output_tokens": null,
  "cost_estimate": null
}
```
