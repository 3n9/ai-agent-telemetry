package dashboard

import (
	"os"
	"strings"
	"testing"

	"github.com/3n9/ai-agent-telemetry/internal/db"
)

func TestResolveInitialRangeDefaultsToLastThirtyDaysOfData(t *testing.T) {
	from, to, err := resolveInitialRange("2026-01-01", "2026-03-11", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if from != "2026-02-10" || to != "2026-03-11" {
		t.Fatalf("unexpected range: got %s to %s", from, to)
	}
}

func TestResolveInitialRangeSupportsAll(t *testing.T) {
	from, to, err := resolveInitialRange("2026-01-01", "2026-03-11", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if from != "2026-01-01" || to != "2026-03-11" {
		t.Fatalf("unexpected range: got %s to %s", from, to)
	}
}

func TestResolveInitialRangeClampsToMinDate(t *testing.T) {
	from, to, err := resolveInitialRange("2026-03-01", "2026-03-11", "30d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if from != "2026-03-01" || to != "2026-03-11" {
		t.Fatalf("unexpected range: got %s to %s", from, to)
	}
}

func TestBuildUsesRequestedInitialRange(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	if err := db.Init(conn); err != nil {
		t.Fatalf("init db: %v", err)
	}

	records := []db.Task{
		{
			ID:                "task-1",
			CreatedAt:         "2026-01-01T10:00:00Z",
			SchemaVersion:     1,
			AgentName:         "codex-cli",
			ModelName:         "gpt-5",
			WorkType:          "analysis",
			WorkTypeTagSource: "recommended",
			Complexity:        "low",
			Confidence:        0.8,
			EstimatedTimeMin:  15,
			TaskType:          "task",
			RawPayloadJSON:    "{}",
		},
		{
			ID:                "task-2",
			CreatedAt:         "2026-03-11T10:00:00Z",
			SchemaVersion:     1,
			AgentName:         "codex-cli",
			ModelName:         "gpt-5",
			WorkType:          "coding",
			WorkTypeTagSource: "recommended",
			Complexity:        "medium",
			Confidence:        0.9,
			EstimatedTimeMin:  30,
			TaskType:          "task",
			RawPayloadJSON:    "{}",
		},
	}

	for _, record := range records {
		record := record
		if err := db.InsertTask(conn, &record); err != nil {
			t.Fatalf("insert task %s: %v", record.ID, err)
		}
	}

	data, err := Build(conn, BuildOptions{InitialRange: "7d"})
	if err != nil {
		t.Fatalf("build dashboard: %v", err)
	}

	if data.InitialFrom != "2026-03-05" || data.InitialTo != "2026-03-11" {
		t.Fatalf("unexpected initial range: got %s to %s", data.InitialFrom, data.InitialTo)
	}
	if data.MinDate != "2026-01-01" || data.MaxDate != "2026-03-11" {
		t.Fatalf("unexpected bounds: got %s to %s", data.MinDate, data.MaxDate)
	}
}

func TestRenderInjectsInitialDates(t *testing.T) {
	html, err := Render(&DashboardData{
		GeneratedAt: "2026-03-11 12:00:00",
		RecordsJSON: "[]",
		MinDate:     "2026-01-01",
		MaxDate:     "2026-03-11",
		InitialFrom: "2026-02-10",
		InitialTo:   "2026-03-11",
	})
	if err != nil {
		t.Fatalf("render dashboard: %v", err)
	}

	content, err := os.ReadFile(html)
	if err != nil {
		t.Fatalf("read rendered html: %v", err)
	}

	page := string(content)
	if !strings.Contains(page, "value=\"2026-02-10\"") || !strings.Contains(page, "value=\"2026-03-11\"") {
		t.Fatalf("expected rendered html to include initial dates")
	}
}
