package db

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func ptr[T any](v T) *T { return &v }

// openInit opens a fresh DB in t.TempDir() and initialises the schema.
func openInit(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("AI_LOG_DB", dbPath)
	conn, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := Init(conn); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return conn
}

// minimalTask returns a Task with every required (NOT NULL) field populated and
// all optional fields as nil/zero.
func minimalTask(id string) *Task {
	return &Task{
		ID:               id,
		CreatedAt:        "2024-01-01T00:00:00Z",
		SchemaVersion:    1,
		AgentName:        "test-agent",
		ModelName:        "test-model",
		WorkType:         "coding",
		WorkTypeTagSource: "recommended",
		Complexity:       "low",
		Confidence:       0.9,
		EstimatedTimeMin: 5,
		TaskType:         "task",
		RawPayloadJSON:   `{}`,
	}
}

// ── DBPath ────────────────────────────────────────────────────────────────────

func TestDBPathEnvVar(t *testing.T) {
	t.Setenv("AI_LOG_DB", "/custom/path/db.sqlite")
	t.Setenv("XDG_DATA_HOME", "")
	got, err := DBPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/custom/path/db.sqlite" {
		t.Errorf("got %q, want /custom/path/db.sqlite", got)
	}
}

func TestDBPathXDGDataHome(t *testing.T) {
	t.Setenv("AI_LOG_DB", "")
	t.Setenv("XDG_DATA_HOME", "/xdg/data")
	got, err := DBPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/xdg/data", xdgSubPath)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDBPathDefault(t *testing.T) {
	t.Setenv("AI_LOG_DB", "")
	t.Setenv("XDG_DATA_HOME", "")
	got, err := DBPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, defaultDBRelPath) {
		t.Errorf("got %q, expected suffix %q", got, defaultDBRelPath)
	}
}

// ── Open ──────────────────────────────────────────────────────────────────────

func TestOpenCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nested", "dir", "test.db")
	t.Setenv("AI_LOG_DB", dbPath)

	conn, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	conn.Close()
}

// ── Init ─────────────────────────────────────────────────────────────────────

func TestInitIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	conn, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	if err := Init(conn); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	if err := Init(conn); err != nil {
		t.Fatalf("second Init: %v", err)
	}
}

func TestInitIntegrityCheck(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	conn, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	if err := Init(conn); err != nil {
		t.Fatalf("Init failed (integrity check may have failed): %v", err)
	}
}

// ── InsertTask ────────────────────────────────────────────────────────────────

func TestInsertTaskAllFields(t *testing.T) {
	conn := openInit(t)

	task := minimalTask("all-fields-1")
	task.SecondaryWorkType = ptr("debugging")
	task.SecondaryWorkTypeTagSource = ptr("custom")
	task.Language = ptr("go")
	task.LanguageTagSource = ptr("recommended")
	task.Domain = ptr("backend")
	task.DomainTagSource = ptr("recommended")
	task.ParentTaskID = ptr("parent-999")
	task.InputTokens = ptr(100)
	task.OutputTokens = ptr(200)
	task.CostEstimate = ptr(0.005)

	if err := InsertTask(conn, task); err != nil {
		t.Fatalf("InsertTask: %v", err)
	}

	var (
		gotID    string
		gotLang  string
		gotCost  float64
	)
	err := conn.QueryRow(
		`SELECT id, language, cost_estimate FROM tasks WHERE id = ?`, task.ID,
	).Scan(&gotID, &gotLang, &gotCost)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if gotID != task.ID {
		t.Errorf("id: got %q, want %q", gotID, task.ID)
	}
	if gotLang != "go" {
		t.Errorf("language: got %q, want go", gotLang)
	}
	if gotCost != 0.005 {
		t.Errorf("cost_estimate: got %v, want 0.005", gotCost)
	}
}

func TestInsertTaskMinimalFields(t *testing.T) {
	conn := openInit(t)
	task := minimalTask("minimal-1")

	if err := InsertTask(conn, task); err != nil {
		t.Fatalf("InsertTask: %v", err)
	}

	var (
		secWork  sql.NullString
		lang     sql.NullString
		domain   sql.NullString
		parent   sql.NullString
		inTok    sql.NullInt64
		outTok   sql.NullInt64
		costEst  sql.NullFloat64
	)
	err := conn.QueryRow(`
		SELECT secondary_work_type, language, domain,
		       parent_task_id, input_tokens, output_tokens, cost_estimate
		FROM tasks WHERE id = ?`, task.ID,
	).Scan(&secWork, &lang, &domain, &parent, &inTok, &outTok, &costEst)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	for name, nv := range map[string]sql.NullString{
		"secondary_work_type": secWork,
		"language":            lang,
		"domain":              domain,
		"parent_task_id":      parent,
	} {
		if nv.Valid {
			t.Errorf("%s: expected NULL, got %q", name, nv.String)
		}
	}
	if inTok.Valid {
		t.Errorf("input_tokens: expected NULL, got %d", inTok.Int64)
	}
	if outTok.Valid {
		t.Errorf("output_tokens: expected NULL, got %d", outTok.Int64)
	}
	if costEst.Valid {
		t.Errorf("cost_estimate: expected NULL, got %v", costEst.Float64)
	}
}

func TestInsertTaskWithTags(t *testing.T) {
	conn := openInit(t)
	task := minimalTask("tagged-1")
	task.Tags = []Tag{
		{Value: "auth", Source: "custom"},
		{Value: "ui", Source: "custom"},
	}

	if err := InsertTask(conn, task); err != nil {
		t.Fatalf("InsertTask: %v", err)
	}

	rows, err := conn.Query(
		`SELECT tag_value, tag_source FROM task_tags WHERE task_id = ? ORDER BY tag_value`,
		task.ID,
	)
	if err != nil {
		t.Fatalf("query tags: %v", err)
	}
	defer rows.Close()

	type row struct{ value, source string }
	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.value, &r.source); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []row{{"auth", "custom"}, {"ui", "custom"}}
	if len(got) != len(want) {
		t.Fatalf("tag count: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tag[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestInsertTaskAtomicity(t *testing.T) {
	conn := openInit(t)

	// Add a unique constraint on (task_id, tag_value) so the second duplicate
	// tag fails mid-transaction, exercising the rollback path.
	if _, err := conn.Exec(
		`CREATE UNIQUE INDEX idx_test_unique_tag ON task_tags(task_id, tag_value)`,
	); err != nil {
		t.Fatalf("create unique index: %v", err)
	}

	task := minimalTask("atomic-1")
	task.Tags = []Tag{
		{Value: "dup", Source: "custom"},
		{Value: "dup", Source: "custom"}, // duplicate → constraint violation
	}

	err := InsertTask(conn, task)
	if err == nil {
		t.Fatal("expected InsertTask to fail due to duplicate tag, got nil")
	}

	// The task row must also have been rolled back.
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM tasks WHERE id = ?`, task.ID).Scan(&count); err != nil {
		t.Fatalf("query tasks: %v", err)
	}
	if count != 0 {
		t.Errorf("task row should have been rolled back, got %d row(s)", count)
	}

	var tagCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM task_tags WHERE task_id = ?`, task.ID).Scan(&tagCount); err != nil {
		t.Fatalf("query task_tags: %v", err)
	}
	if tagCount != 0 {
		t.Errorf("tag rows should have been rolled back, got %d row(s)", tagCount)
	}
}

// ── Reset ─────────────────────────────────────────────────────────────────────

func TestResetClearsData(t *testing.T) {
	conn := openInit(t)

	task := minimalTask("reset-1")
	task.Tags = []Tag{{Value: "x", Source: "custom"}}
	if err := InsertTask(conn, task); err != nil {
		t.Fatalf("InsertTask: %v", err)
	}

	if err := Reset(conn); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	var taskCount, tagCount int
	conn.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&taskCount)
	conn.QueryRow(`SELECT COUNT(*) FROM task_tags`).Scan(&tagCount)

	if taskCount != 0 {
		t.Errorf("tasks: got %d rows after Reset, want 0", taskCount)
	}
	if tagCount != 0 {
		t.Errorf("task_tags: got %d rows after Reset, want 0", tagCount)
	}
}

// ── task_parent_status view ───────────────────────────────────────────────────

func TestParentLinkStatusView(t *testing.T) {
	conn := openInit(t)

	parent := minimalTask("parent-1")
	child := minimalTask("child-1")
	child.ParentTaskID = ptr("parent-1")
	orphan := minimalTask("orphan-1")
	orphan.ParentTaskID = ptr("nonexistent-parent")
	noParent := minimalTask("no-parent-1")

	for _, task := range []*Task{parent, child, orphan, noParent} {
		if err := InsertTask(conn, task); err != nil {
			t.Fatalf("InsertTask(%s): %v", task.ID, err)
		}
	}

	rows, err := conn.Query(
		`SELECT id, parent_link_status FROM task_parent_status ORDER BY id`,
	)
	if err != nil {
		t.Fatalf("query view: %v", err)
	}
	defer rows.Close()

	statuses := map[string]string{}
	for rows.Next() {
		var id, status string
		if err := rows.Scan(&id, &status); err != nil {
			t.Fatal(err)
		}
		statuses[id] = status
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	cases := map[string]string{
		"parent-1":   "none",
		"child-1":    "linked",
		"orphan-1":   "dangling",
		"no-parent-1": "none",
	}
	for id, want := range cases {
		got, ok := statuses[id]
		if !ok {
			t.Errorf("id %q not found in view", id)
			continue
		}
		if got != want {
			t.Errorf("id %q: parent_link_status = %q, want %q", id, got, want)
		}
	}
}
