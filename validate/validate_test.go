package validate

import (
	"fmt"
	"strings"
	"testing"
)

func validPayloadJSON() string {
	return `{"schema_version":1,"agent_name":"copilot-cli","model_name":"claude-3-5-sonnet","work_type":"coding","complexity":"medium","confidence":0.9,"estimated_time_min":10,"task_type":"task"}`
}

func basePayload() *Payload {
	return &Payload{
		SchemaVersion:    1,
		AgentName:        "a",
		ModelName:        "m",
		WorkType:         "coding",
		Complexity:       "low",
		Confidence:       0.5,
		EstimatedTimeMin: 10,
		TaskType:         "task",
	}
}

func TestParse_ValidPayload(t *testing.T) {
	p, err := Parse(validPayloadJSON())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil payload")
	}
}

func TestCheck_ValidPayload(t *testing.T) {
	p, parseErr := Parse(validPayloadJSON())
	if parseErr != nil {
		t.Fatalf("parse failed: %v", parseErr)
	}
	if err := Check(p); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestCheck_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Payload)
		wantMsg string
	}{
		{
			name:    "blank agent_name",
			mutate:  func(p *Payload) { p.AgentName = "" },
			wantMsg: "agent_name is required",
		},
		{
			name:    "whitespace agent_name",
			mutate:  func(p *Payload) { p.AgentName = "   " },
			wantMsg: "agent_name is required",
		},
		{
			name:    "blank model_name",
			mutate:  func(p *Payload) { p.ModelName = "" },
			wantMsg: "model_name is required",
		},
		{
			name:    "blank work_type",
			mutate:  func(p *Payload) { p.WorkType = "" },
			wantMsg: "work_type is required",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := basePayload()
			tc.mutate(p)
			err := Check(p)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Message, tc.wantMsg) {
				t.Errorf("want message containing %q, got %q", tc.wantMsg, err.Message)
			}
		})
	}
}

func TestCheck_SchemaVersion(t *testing.T) {
	tests := []struct {
		version int
		wantErr bool
	}{
		{1, false},
		{0, true},
		{2, true},
		{-1, true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("version_%d", tc.version), func(t *testing.T) {
			p := basePayload()
			p.SchemaVersion = tc.version
			err := Check(p)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestCheck_Complexity(t *testing.T) {
	tests := []struct {
		complexity string
		wantErr    bool
	}{
		{"low", false},
		{"medium", false},
		{"high", false},
		{"extreme", true},
		{"", true},
		{"Low", true},
		{"MEDIUM", true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("complexity_%q", tc.complexity), func(t *testing.T) {
			p := basePayload()
			p.Complexity = tc.complexity
			err := Check(p)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestCheck_Confidence(t *testing.T) {
	tests := []struct {
		confidence float64
		wantErr    bool
	}{
		{0.0, false},
		{1.0, false},
		{0.5, false},
		{-0.1, true},
		{1.1, true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("confidence_%v", tc.confidence), func(t *testing.T) {
			p := basePayload()
			p.Confidence = tc.confidence
			err := Check(p)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestCheck_EstimatedTimeMin(t *testing.T) {
	tests := []struct {
		minutes int
		wantErr bool
	}{
		{1, false},
		{240, false},
		{60, false},
		{0, true},
		{241, true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("minutes_%d", tc.minutes), func(t *testing.T) {
			p := basePayload()
			p.EstimatedTimeMin = tc.minutes
			err := Check(p)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestCheck_TaskType(t *testing.T) {
	tests := []struct {
		taskType string
		wantErr  bool
	}{
		{"task", false},
		{"subtask", false},
		{"interruption", false},
		{"other", true},
		{"", true},
		{"Task", true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("task_type_%q", tc.taskType), func(t *testing.T) {
			p := basePayload()
			p.TaskType = tc.taskType
			err := Check(p)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestCheck_CustomTags_MaxItems(t *testing.T) {
	p := basePayload()
	p.CustomTags = []string{"a", "b", "c", "d", "e"}
	if err := Check(p); err != nil {
		t.Errorf("5 tags should pass, got: %v", err)
	}

	p2 := basePayload()
	p2.CustomTags = []string{"a", "b", "c", "d", "e", "f"}
	if err := Check(p2); err == nil {
		t.Error("6 tags should fail")
	} else if !strings.Contains(err.Message, "at most 5") {
		t.Errorf("unexpected error message: %q", err.Message)
	}
}

func TestCheck_CustomTags_ItemLength(t *testing.T) {
	p := basePayload()
	p.CustomTags = []string{strings.Repeat("a", 32)}
	if err := Check(p); err != nil {
		t.Errorf("tag of 32 chars should pass, got: %v", err)
	}

	p2 := basePayload()
	p2.CustomTags = []string{strings.Repeat("a", 33)}
	if err := Check(p2); err == nil {
		t.Error("tag of 33 chars should fail")
	} else if !strings.Contains(err.Message, "32 characters or fewer") {
		t.Errorf("unexpected error message: %q", err.Message)
	}
}

func TestCheck_StringLengthLimits(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Payload)
		wantErr bool
	}{
		{"agent_name 64 chars", func(p *Payload) { p.AgentName = strings.Repeat("a", 64) }, false},
		{"agent_name 65 chars", func(p *Payload) { p.AgentName = strings.Repeat("a", 65) }, true},
		{"model_name 128 chars", func(p *Payload) { p.ModelName = strings.Repeat("m", 128) }, false},
		{"model_name 129 chars", func(p *Payload) { p.ModelName = strings.Repeat("m", 129) }, true},
		{"work_type 32 chars", func(p *Payload) { p.WorkType = strings.Repeat("w", 32) }, false},
		{"work_type 33 chars", func(p *Payload) { p.WorkType = strings.Repeat("w", 33) }, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := basePayload()
			tc.mutate(p)
			err := Check(p)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestParse_UnknownFields(t *testing.T) {
	raw := `{"schema_version":1,"agent_name":"a","model_name":"m","work_type":"coding","complexity":"low","confidence":0.5,"estimated_time_min":10,"task_type":"task","unknown_field":"value"}`
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
	if err.Code != "PARSE_ERROR" {
		t.Errorf("expected code PARSE_ERROR, got: %s", err.Code)
	}
}

func TestCheck_OptionalFieldsOmitted(t *testing.T) {
	p, parseErr := Parse(validPayloadJSON())
	if parseErr != nil {
		t.Fatalf("parse failed: %v", parseErr)
	}
	if p.Language != nil {
		t.Error("language should be nil when omitted")
	}
	if p.Domain != nil {
		t.Error("domain should be nil when omitted")
	}
	if p.SecondaryWorkType != nil {
		t.Error("secondary_work_type should be nil when omitted")
	}
	if p.ParentTaskID != nil {
		t.Error("parent_task_id should be nil when omitted")
	}
	if err := Check(p); err != nil {
		t.Errorf("optional fields omitted should pass, got: %v", err)
	}
}

func TestCheck_ParentTaskID(t *testing.T) {
	ptr := func(s string) *string { return &s }

	tests := []struct {
		name    string
		id      *string
		wantErr bool
	}{
		{"nil (omitted)", nil, false},
		{"short id", ptr("abc"), false},
		{"64 chars", ptr(strings.Repeat("x", 64)), false},
		{"65 chars", ptr(strings.Repeat("x", 65)), true},
		{"empty string", ptr(""), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := basePayload()
			p.ParentTaskID = tc.id
			err := Check(p)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}
