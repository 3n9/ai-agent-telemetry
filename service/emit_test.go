package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/3n9/ai-agent-telemetry/internal/db"
	"github.com/3n9/ai-agent-telemetry/validate"
)

func TestEmitStoresTaskAndNormalizesTags(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "telemetry.db")

	prev := os.Getenv("AI_LOG_DB")
	if err := os.Setenv("AI_LOG_DB", dbPath); err != nil {
		t.Fatalf("set AI_LOG_DB: %v", err)
	}
	defer func() {
		if prev == "" {
			_ = os.Unsetenv("AI_LOG_DB")
			return
		}
		_ = os.Setenv("AI_LOG_DB", prev)
	}()

	payload := &validate.Payload{
		SchemaVersion:    1,
		AgentName:        "codex-cli",
		ModelName:        "gpt-5",
		WorkType:         "coding",
		Complexity:       "medium",
		Confidence:       0.8,
		EstimatedTimeMin: 15,
		TaskType:         "task",
		CustomTags:       []string{"UI Polish", "auth"},
	}

	resp, emitErr := Emit(EmitRequest{Payload: payload})
	if emitErr != nil {
		t.Fatalf("Emit returned error: %v", emitErr)
	}
	if !resp.OK {
		t.Fatalf("expected OK response")
	}
	if resp.TaskID == "" {
		t.Fatalf("expected generated task ID")
	}
	if len(resp.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(resp.Warnings))
	}
	if got := resp.Warnings[0]["to"]; got != "ui-polish" {
		t.Fatalf("expected normalized tag ui-polish, got %q", got)
	}

	conn, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count); err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 stored task, got %d", count)
	}

	var raw string
	if err := conn.QueryRow(`SELECT raw_payload_json FROM tasks LIMIT 1`).Scan(&raw); err != nil {
		t.Fatalf("load raw payload: %v", err)
	}
	if raw == "" {
		t.Fatalf("expected raw payload json to be stored")
	}
}
