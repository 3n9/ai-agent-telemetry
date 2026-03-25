package db

// Tag represents a stored task tag.
type Tag struct {
	Value  string
	Source string // "recommended" or "custom"
}

// Task is the full record written to the tasks table.
type Task struct {
	ID                         string
	CreatedAt                  string
	SchemaVersion              int
	AgentName                  string
	ModelName                  string
	WorkType                   string
	WorkTypeTagSource          string
	SecondaryWorkType          *string
	SecondaryWorkTypeTagSource *string
	Language                   *string
	LanguageTagSource          *string
	Domain                     *string
	DomainTagSource            *string
	Complexity                 string
	Confidence                 float64
	EstimatedTimeMin           int
	TaskType                   string
	ParentTaskID               *string
	InputTokens                *int
	OutputTokens               *int
	CostEstimate               *float64
	RawPayloadJSON             string
	Tags                       []Tag
}
