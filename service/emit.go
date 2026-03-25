package service

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/3n9/ai-agent-telemetry/internal/db"
	"github.com/3n9/ai-agent-telemetry/internal/tags"
	"github.com/3n9/ai-agent-telemetry/internal/ulidgen"
	"github.com/3n9/ai-agent-telemetry/validate"
)

// Error is a structured service failure.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

// EmitRequest describes one telemetry record to store.
type EmitRequest struct {
	Payload            *validate.Payload
	RawPayloadJSON     string
	TaskIDOverride     *string
	ParentTaskOverride *string
}

// EmitResponse mirrors the CLI success payload.
type EmitResponse struct {
	OK            bool                `json:"ok"`
	TaskID        string              `json:"task_id"`
	ParentTaskID  *string             `json:"parent_task_id,omitempty"`
	TaskType      string              `json:"task_type"`
	SchemaVersion int                 `json:"schema_version"`
	Warnings      []map[string]string `json:"warnings"`
}

// ParseAndValidate parses JSON and applies business validation.
func ParseAndValidate(raw string) (*validate.Payload, *Error) {
	p, verr := validate.Parse(raw)
	if verr != nil {
		slog.Default().Warn("validation error", "code", verr.Code, "message", verr.Message)
		return nil, &Error{Code: verr.Code, Message: verr.Message}
	}
	if verr := validate.Check(p); verr != nil {
		slog.Default().Warn("validation error", "code", verr.Code, "message", verr.Message)
		return nil, &Error{Code: verr.Code, Message: verr.Message}
	}
	return p, nil
}

// Emit validates, stores, and returns the stored task metadata.
func Emit(req EmitRequest) (*EmitResponse, *Error) {
	p := req.Payload
	if p == nil {
		if req.RawPayloadJSON == "" {
			return nil, &Error{Code: "VALIDATION_ERROR", Message: "payload is required"}
		}
		var err *Error
		p, err = ParseAndValidate(req.RawPayloadJSON)
		if err != nil {
			return nil, err
		}
	}

	copyPayload := *p
	p = &copyPayload

	if req.TaskIDOverride != nil {
		p.TaskID = req.TaskIDOverride
	}
	if req.ParentTaskOverride != nil {
		p.ParentTaskID = req.ParentTaskOverride
	}
	if verr := validate.Check(p); verr != nil {
		slog.Default().Warn("validation error", "code", verr.Code, "message", verr.Message)
		return nil, &Error{Code: verr.Code, Message: verr.Message}
	}

	raw := req.RawPayloadJSON
	if raw == "" || req.TaskIDOverride != nil || req.ParentTaskOverride != nil {
		marshaled, err := json.Marshal(p)
		if err != nil {
			return nil, &Error{Code: "PARSE_ERROR", Message: err.Error()}
		}
		raw = string(marshaled)
	}

	id := ulidgen.New()
	if p.TaskID != nil && *p.TaskID != "" {
		id = *p.TaskID
	}

	workTypeSource := tags.Source(p.WorkType, tags.RecommendedWorkTypes)
	var secSource *string
	if p.SecondaryWorkType != nil {
		s := tags.Source(*p.SecondaryWorkType, tags.RecommendedWorkTypes)
		secSource = &s
	}
	var langSource *string
	if p.Language != nil {
		s := tags.Source(*p.Language, tags.RecommendedLanguages)
		langSource = &s
	}
	var domainSource *string
	if p.Domain != nil {
		s := tags.Source(*p.Domain, tags.RecommendedDomains)
		domainSource = &s
	}

	normalised := tags.NormalizeAll(p.CustomTags)
	tagRows := make([]db.Tag, 0, len(normalised))
	for _, t := range normalised {
		tagRows = append(tagRows, db.Tag{Value: t, Source: "custom"})
	}

	var warnings []map[string]string
	for i, orig := range p.CustomTags {
		if i < len(normalised) && orig != normalised[i] {
			warnings = append(warnings, map[string]string{
				"code":    "TAG_NORMALIZED",
				"message": "tag was normalized",
				"field":   "custom_tags",
				"from":    orig,
				"to":      normalised[i],
			})
		}
	}
	if warnings == nil {
		warnings = []map[string]string{}
	}

	task := &db.Task{
		ID:                         id,
		CreatedAt:                  time.Now().UTC().Format(time.RFC3339),
		SchemaVersion:              p.SchemaVersion,
		AgentName:                  p.AgentName,
		ModelName:                  p.ModelName,
		WorkType:                   p.WorkType,
		WorkTypeTagSource:          workTypeSource,
		SecondaryWorkType:          p.SecondaryWorkType,
		SecondaryWorkTypeTagSource: secSource,
		Language:                   p.Language,
		LanguageTagSource:          langSource,
		Domain:                     p.Domain,
		DomainTagSource:            domainSource,
		Complexity:                 p.Complexity,
		Confidence:                 p.Confidence,
		EstimatedTimeMin:           p.EstimatedTimeMin,
		TaskType:                   p.TaskType,
		ParentTaskID:               p.ParentTaskID,
		InputTokens:                p.InputTokens,
		OutputTokens:               p.OutputTokens,
		CostEstimate:               p.CostEstimate,
		RawPayloadJSON:             raw,
		Tags:                       tagRows,
	}

	path, err := db.DBPath()
	if err != nil {
		slog.Default().Error("DB path error", "code", "DB_PATH_ERROR", "error", err)
		return nil, &Error{Code: "DB_PATH_ERROR", Message: err.Error()}
	}
	conn, err := db.Open(path)
	if err != nil {
		slog.Default().Error("DB open error", "code", "DB_OPEN_ERROR", "error", err)
		return nil, &Error{Code: "DB_OPEN_ERROR", Message: err.Error()}
	}
	defer conn.Close()

	if err := db.Init(conn); err != nil {
		slog.Default().Error("DB init error", "code", "DB_INIT_ERROR", "error", err)
		return nil, &Error{Code: "DB_INIT_ERROR", Message: err.Error()}
	}
	if err := db.InsertTask(conn, task); err != nil {
		dbErr := classifyDBError(err)
		slog.Default().Error("DB write error", "code", dbErr.Code, "error", err)
		return nil, dbErr
	}

	slog.Default().Debug("emit succeeded", "task_id", id, "task_type", p.TaskType, "agent_name", p.AgentName)
	return &EmitResponse{
		OK:            true,
		TaskID:        id,
		ParentTaskID:  p.ParentTaskID,
		TaskType:      p.TaskType,
		SchemaVersion: p.SchemaVersion,
		Warnings:      warnings,
	}, nil
}

func classifyDBError(err error) *Error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "permission denied"):
		return &Error{Code: "DB_PERMISSION_ERROR", Message: msg}
	case strings.Contains(msg, "no space left") || strings.Contains(msg, "disk full"):
		return &Error{Code: "DB_DISK_FULL_ERROR", Message: msg}
	case strings.Contains(msg, "database is locked") || strings.Contains(msg, "SQLITE_BUSY"):
		return &Error{Code: "DB_LOCKED_ERROR", Message: msg}
	default:
		return &Error{Code: "DB_WRITE_ERROR", Message: msg}
	}
}
