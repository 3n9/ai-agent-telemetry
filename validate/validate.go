package validate

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Payload mirrors the ai-log JSON payload schema.
type Payload struct {
	SchemaVersion    int      `json:"schema_version"`
	TaskID           *string  `json:"task_id,omitempty"`
	AgentName        string   `json:"agent_name"`
	ModelName        string   `json:"model_name"`
	WorkType         string   `json:"work_type"`
	SecondaryWorkType *string `json:"secondary_work_type"`
	Language         *string  `json:"language"`
	Domain           *string  `json:"domain"`
	CustomTags       []string `json:"custom_tags"`
	Complexity       string   `json:"complexity"`
	Confidence       float64  `json:"confidence"`
	EstimatedTimeMin int      `json:"estimated_time_min"`
	TaskType         string   `json:"task_type"`
	ParentTaskID     *string  `json:"parent_task_id"`
	InputTokens      *int     `json:"input_tokens"`
	OutputTokens     *int     `json:"output_tokens"`
	CostEstimate     *float64 `json:"cost_estimate"`
}

// ValidationError holds a structured validation failure.
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

var (
	validComplexity = map[string]bool{"low": true, "medium": true, "high": true}
	validTaskTypes  = map[string]bool{"task": true, "subtask": true, "interruption": true}
)

// Parse parses raw JSON into a Payload, returning a ValidationError on failure.
func Parse(raw string) (*Payload, *ValidationError) {
	var p Payload
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return nil, &ValidationError{Code: "PARSE_ERROR", Message: err.Error()}
	}
	return &p, nil
}

// Check validates business rules on an already-parsed Payload.
func Check(p *Payload) *ValidationError {
	if p.SchemaVersion != 1 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "schema_version must be 1"}
	}
	if strings.TrimSpace(p.AgentName) == "" {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "agent_name is required"}
	}
	if len(p.AgentName) > 64 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "agent_name must be 64 characters or fewer"}
	}
	if strings.TrimSpace(p.ModelName) == "" {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "model_name is required"}
	}
	if len(p.ModelName) > 128 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "model_name must be 128 characters or fewer"}
	}
	if strings.TrimSpace(p.WorkType) == "" {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "work_type is required"}
	}
	if len(p.WorkType) > 32 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "work_type must be 32 characters or fewer"}
	}
	if !validComplexity[p.Complexity] {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "complexity must be one of: low, medium, high"}
	}
	if p.Confidence < 0 || p.Confidence > 1 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "confidence must be between 0 and 1"}
	}
	if p.EstimatedTimeMin < 1 || p.EstimatedTimeMin > 240 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "estimated_time_min must be between 1 and 240"}
	}
	if !validTaskTypes[p.TaskType] {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "task_type must be one of: task, subtask, interruption"}
	}
	if len(p.CustomTags) > 5 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "custom_tags may have at most 5 items"}
	}
	for i, tag := range p.CustomTags {
		if len(tag) > 32 {
			return &ValidationError{Code: "VALIDATION_ERROR", Message: fmt.Sprintf("custom_tags[%d] must be 32 characters or fewer", i)}
		}
	}
	if p.TaskID != nil && (len(*p.TaskID) < 1 || len(*p.TaskID) > 64) {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "task_id must be between 1 and 64 characters"}
	}
	if p.ParentTaskID != nil && (len(*p.ParentTaskID) < 1 || len(*p.ParentTaskID) > 64) {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "parent_task_id must be between 1 and 64 characters"}
	}
	if p.InputTokens != nil && *p.InputTokens < 0 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "input_tokens must be >= 0"}
	}
	if p.OutputTokens != nil && *p.OutputTokens < 0 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "output_tokens must be >= 0"}
	}
	if p.CostEstimate != nil && *p.CostEstimate < 0 {
		return &ValidationError{Code: "VALIDATION_ERROR", Message: "cost_estimate must be >= 0"}
	}
	return nil
}
