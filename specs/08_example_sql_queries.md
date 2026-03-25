# Example SQL Queries

These queries assume the MVP schema described in `05_technical_spec.md`.

## Total estimated time by work type

```sql
SELECT work_type, SUM(estimated_time_min) AS total_minutes
FROM tasks
WHERE task_type = 'task'
GROUP BY work_type
ORDER BY total_minutes DESC;
```

## Task counts by model

```sql
SELECT model_name, COUNT(*) AS task_count
FROM tasks
WHERE task_type = 'task'
GROUP BY model_name
ORDER BY task_count DESC;
```

## Interruptions by model

```sql
SELECT model_name, COUNT(*) AS interruption_count
FROM tasks
WHERE task_type = 'interruption'
GROUP BY model_name
ORDER BY interruption_count DESC;
```

## Average confidence by work type

```sql
SELECT work_type, ROUND(AVG(confidence), 3) AS avg_confidence
FROM tasks
GROUP BY work_type
ORDER BY avg_confidence DESC;
```

## Complexity distribution

```sql
SELECT complexity, COUNT(*) AS count
FROM tasks
GROUP BY complexity
ORDER BY CASE complexity
    WHEN 'low' THEN 1
    WHEN 'medium' THEN 2
    WHEN 'high' THEN 3
END;
```

## Recommended vs custom work type usage

```sql
SELECT work_type_tag_source, COUNT(*) AS count
FROM tasks
GROUP BY work_type_tag_source
ORDER BY count DESC;
```

## Top custom tags

```sql
SELECT tag_value, COUNT(*) AS count
FROM task_tags
GROUP BY tag_value
ORDER BY count DESC
LIMIT 20;
```

## Null language/domain rate

```sql
SELECT
  SUM(CASE WHEN language IS NULL THEN 1 ELSE 0 END) AS null_language_count,
  SUM(CASE WHEN domain IS NULL THEN 1 ELSE 0 END) AS null_domain_count,
  COUNT(*) AS total_rows
FROM tasks;
```

## Parent link status summary

```sql
SELECT parent_link_status, COUNT(*) AS count
FROM task_parent_status
GROUP BY parent_link_status
ORDER BY count DESC;
```

## Estimated time by model and complexity

```sql
SELECT model_name, complexity, SUM(estimated_time_min) AS total_minutes
FROM tasks
GROUP BY model_name, complexity
ORDER BY model_name, complexity;
```

## Attached vs standalone interruptions

```sql
SELECT
  CASE
    WHEN parent_task_id IS NULL THEN 'standalone'
    ELSE 'attached'
  END AS interruption_kind,
  COUNT(*) AS count
FROM tasks
WHERE task_type = 'interruption'
GROUP BY interruption_kind;
```

## Top work types within a specific model

```sql
SELECT work_type, COUNT(*) AS count
FROM tasks
WHERE model_name = 'gpt-5'
GROUP BY work_type
ORDER BY count DESC;
```
