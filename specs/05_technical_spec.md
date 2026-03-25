# Technical Specification

## Recommended Vocabularies

### work_type
- coding
- debugging
- research
- analysis
- writing
- planning
- creative
- support
- refactor

### language
- php
- javascript
- typescript
- python
- sql
- html
- css
- shell
- json
- yaml
- markdown
- none

### domain
- frontend
- backend
- database
- devops
- documentation
- wordpress
- laravel
- api
- testing
- fiction
- horror
- email
- blog
- marketing
- none

These starter vocabularies are advisory in v1. They are used to derive `tag_source`.

---

## Payload Schema

Required fields:

- schema_version
- agent_name
- model_name
- work_type
- complexity
- confidence
- estimated_time_min
- task_type

Optional fields:

- secondary_work_type
- language
- domain
- custom_tags
- parent_task_id
- input_tokens
- output_tokens
- cost_estimate

---

## Example Payload (Creative Work)

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
  "task_type": "task"
}
```

## Example Payload (Writing Work)

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
  "task_type": "task"
}
```

---

## Database Schema

### tasks

```sql
CREATE TABLE tasks (
 id TEXT PRIMARY KEY,
 created_at TEXT NOT NULL,
 schema_version INTEGER NOT NULL,
 agent_name TEXT NOT NULL,
 model_name TEXT NOT NULL,
 work_type TEXT NOT NULL,
 work_type_tag_source TEXT NOT NULL,
 secondary_work_type TEXT,
 secondary_work_type_tag_source TEXT,
 language TEXT,
 language_tag_source TEXT,
 domain TEXT,
 domain_tag_source TEXT,
 complexity TEXT NOT NULL,
 confidence REAL NOT NULL,
 estimated_time_min INTEGER NOT NULL,
 task_type TEXT NOT NULL,
 parent_task_id TEXT,
 input_tokens INTEGER,
 output_tokens INTEGER,
 cost_estimate REAL,
 raw_payload_json TEXT NOT NULL
);
```

### task_tags

```sql
CREATE TABLE task_tags (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 task_id TEXT NOT NULL,
 tag_value TEXT NOT NULL,
 tag_source TEXT NOT NULL
);
```

---

## Views

### Parent link status

```sql
CREATE VIEW task_parent_status AS
SELECT
 t.*,
 CASE
  WHEN parent_task_id IS NULL THEN 'none'
  WHEN EXISTS (
   SELECT 1 FROM tasks p WHERE p.id = t.parent_task_id
  ) THEN 'linked'
  ELSE 'dangling'
 END AS parent_link_status
FROM tasks t;
```

---

## Validation Rules

- `confidence` must be between 0 and 1
- `estimated_time_min` must be between 1 and 240
- `complexity` must be one of `low`, `medium`, `high`
- `task_type` must be one of `task`, `subtask`, `interruption`
- `language` and `domain` may be null
- interruptions may have `parent_task_id`
- subtasks may have dangling `parent_task_id`

---

## Indexes

```sql
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_work_type ON tasks(work_type);
CREATE INDEX idx_tasks_model_name ON tasks(model_name);
CREATE INDEX idx_tasks_complexity ON tasks(complexity);
CREATE INDEX idx_tasks_task_type ON tasks(task_type);
CREATE INDEX idx_task_tags_tag_value ON task_tags(tag_value);
```
