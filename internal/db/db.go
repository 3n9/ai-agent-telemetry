package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const defaultDBRelPath = ".local/ai-telemetry/telemetry.db"
const xdgSubPath = "ai-telemetry/telemetry.db"

// DBPath returns the database file path, respecting AI_LOG_DB, XDG_DATA_HOME, or the
// legacy default (~/.local/ai-telemetry/telemetry.db) for existing installs.
func DBPath() (string, error) {
	if p := os.Getenv("AI_LOG_DB"); p != "" {
		return p, nil
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, xdgSubPath), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, defaultDBRelPath), nil
}

// Open opens (and optionally creates) the SQLite database at the given path.
func Open(path string) (*sql.DB, error) {
	slog.Default().Debug("opening database", "path", path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("cannot create database directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("cannot set pragmas: %w", err)
	}
	return db, nil
}

// Init creates tables, views, and indexes if they do not already exist.
func Init(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS tasks (
  id                              TEXT PRIMARY KEY,
  created_at                      TEXT NOT NULL,
  schema_version                  INTEGER NOT NULL,
  agent_name                      TEXT NOT NULL,
  model_name                      TEXT NOT NULL,
  work_type                       TEXT NOT NULL,
  work_type_tag_source            TEXT NOT NULL,
  secondary_work_type             TEXT,
  secondary_work_type_tag_source  TEXT,
  language                        TEXT,
  language_tag_source             TEXT,
  domain                          TEXT,
  domain_tag_source               TEXT,
  complexity                      TEXT NOT NULL,
  confidence                      REAL NOT NULL,
  estimated_time_min              INTEGER NOT NULL,
  task_type                       TEXT NOT NULL,
  parent_task_id                  TEXT,
  input_tokens                    INTEGER,
  output_tokens                   INTEGER,
  cost_estimate                   REAL,
  raw_payload_json                TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS task_tags (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  task_id     TEXT NOT NULL,
  tag_value   TEXT NOT NULL,
  tag_source  TEXT NOT NULL
);

CREATE VIEW IF NOT EXISTS task_parent_status AS
SELECT
  t.*,
  CASE
    WHEN t.parent_task_id IS NULL THEN 'none'
    WHEN EXISTS (SELECT 1 FROM tasks p WHERE p.id = t.parent_task_id) THEN 'linked'
    ELSE 'dangling'
  END AS parent_link_status
FROM tasks t;

CREATE INDEX IF NOT EXISTS idx_tasks_created_at   ON tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_work_type     ON tasks(work_type);
CREATE INDEX IF NOT EXISTS idx_tasks_model_name    ON tasks(model_name);
CREATE INDEX IF NOT EXISTS idx_tasks_complexity    ON tasks(complexity);
CREATE INDEX IF NOT EXISTS idx_tasks_task_type     ON tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_task_tags_tag_value ON task_tags(tag_value);
`)
	if err != nil {
		return fmt.Errorf("schema init failed: %w", err)
	}
	var integrityResult string
	if err := db.QueryRow("PRAGMA integrity_check").Scan(&integrityResult); err != nil {
		return fmt.Errorf("integrity_check failed: %w", err)
	}
	if integrityResult != "ok" {
		slog.Default().Warn("database integrity check failed", "result", integrityResult)
		return fmt.Errorf("database integrity check failed: %s", integrityResult)
	}
	slog.Default().Debug("database schema ready")
	return nil
}

// Reset deletes all rows from tasks and task_tags, preserving the schema.
func Reset(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM task_tags; DELETE FROM tasks;`)
	if err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}
	return nil
}
// InsertTask inserts a task row
func InsertTask(db *sql.DB, t *Task) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
INSERT INTO tasks (
  id, created_at, schema_version, agent_name, model_name,
  work_type, work_type_tag_source,
  secondary_work_type, secondary_work_type_tag_source,
  language, language_tag_source,
  domain, domain_tag_source,
  complexity, confidence, estimated_time_min,
  task_type, parent_task_id,
  input_tokens, output_tokens, cost_estimate,
  raw_payload_json
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.CreatedAt, t.SchemaVersion,
		t.AgentName, t.ModelName,
		t.WorkType, t.WorkTypeTagSource,
		t.SecondaryWorkType, t.SecondaryWorkTypeTagSource,
		t.Language, t.LanguageTagSource,
		t.Domain, t.DomainTagSource,
		t.Complexity, t.Confidence, t.EstimatedTimeMin,
		t.TaskType, t.ParentTaskID,
		t.InputTokens, t.OutputTokens, t.CostEstimate,
		t.RawPayloadJSON,
	)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	for _, tag := range t.Tags {
		if _, err := tx.Exec(
			`INSERT INTO task_tags (task_id, tag_value, tag_source) VALUES (?,?,?)`,
			t.ID, tag.Value, tag.Source,
		); err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	slog.Default().Debug("task inserted", "task_id", t.ID)
	return nil
}
