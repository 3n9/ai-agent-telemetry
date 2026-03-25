package dashboard

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//go:embed static/echarts.min.js
var echartsJS string

// ── Data structures ───────────────────────────────────────────────────────────

type RecordTag struct {
	Value  string `json:"value"`
	Source string `json:"source"`
}

type Record struct {
	ID               string      `json:"id"`
	CreatedAt        string      `json:"created_at"`
	AgentName        string      `json:"agent_name"`
	ModelName        string      `json:"model_name"`
	WorkType         string      `json:"work_type"`
	WorkTypeSource   string      `json:"work_type_tag_source"`
	SecWorkType      string      `json:"secondary_work_type"`
	Language         string      `json:"language"`
	Domain           string      `json:"domain"`
	Complexity       string      `json:"complexity"`
	Confidence       float64     `json:"confidence"`
	EstimatedMin     int         `json:"estimated_time_min"`
	TaskType         string      `json:"task_type"`
	ParentTaskID     string      `json:"parent_task_id"`
	ParentLinkStatus string      `json:"parent_link_status"`
	Tags             []RecordTag `json:"tags"`
}

type DashboardData struct {
	GeneratedAt string
	RecordsJSON string
	MinDate     string
	MaxDate     string
	InitialFrom string
	InitialTo   string
}

type BuildOptions struct {
	InitialRange string
}

// ── Build ─────────────────────────────────────────────────────────────────────

func Build(db *sql.DB, opts BuildOptions) (*DashboardData, error) {
	tagRows, err := db.Query(`SELECT task_id, tag_value, tag_source FROM task_tags`)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer tagRows.Close()

	tagMap := map[string][]RecordTag{}
	for tagRows.Next() {
		var tid, val, src string
		if err := tagRows.Scan(&tid, &val, &src); err != nil {
			return nil, fmt.Errorf("scan tag row: %w", err)
		}
		tagMap[tid] = append(tagMap[tid], RecordTag{Value: val, Source: src})
	}
	if err := tagRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tag rows: %w", err)
	}

	rows, err := db.Query(`
SELECT
id, created_at, agent_name, model_name,
work_type, COALESCE(work_type_tag_source,'custom'),
COALESCE(secondary_work_type,''), COALESCE(language,''), COALESCE(domain,''),
complexity, confidence, estimated_time_min,
task_type, COALESCE(parent_task_id,''), parent_link_status
FROM task_parent_status
ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var records []Record
	var minDate, maxDate string

	for rows.Next() {
		var r Record
		if err := rows.Scan(
			&r.ID, &r.CreatedAt, &r.AgentName, &r.ModelName,
			&r.WorkType, &r.WorkTypeSource,
			&r.SecWorkType, &r.Language, &r.Domain,
			&r.Complexity, &r.Confidence, &r.EstimatedMin,
			&r.TaskType, &r.ParentTaskID, &r.ParentLinkStatus,
		); err != nil {
			return nil, fmt.Errorf("scan task row: %w", err)
		}
		r.Tags = tagMap[r.ID]
		if r.Tags == nil {
			r.Tags = []RecordTag{}
		}

		day := dayOnly(r.CreatedAt)
		if minDate == "" || day < minDate {
			minDate = day
		}
		if maxDate == "" || day > maxDate {
			maxDate = day
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate task rows: %w", err)
	}

	if records == nil {
		records = []Record{}
	}

	initialFrom, initialTo, err := resolveInitialRange(minDate, maxDate, opts.InitialRange)
	if err != nil {
		return nil, fmt.Errorf("resolve initial range: %w", err)
	}

	b, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("marshal records: %w", err)
	}

	return &DashboardData{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		RecordsJSON: string(b),
		MinDate:     minDate,
		MaxDate:     maxDate,
		InitialFrom: initialFrom,
		InitialTo:   initialTo,
	}, nil
}

// ── Render ────────────────────────────────────────────────────────────────────

func Render(data *DashboardData) (string, error) {
	html := strings.Replace(htmlTemplate, "/*__ECHARTS_PLACEHOLDER__*/", echartsJS, 1)
	html = strings.ReplaceAll(html, "__MIN_DATE__", data.MinDate)
	html = strings.ReplaceAll(html, "__MAX_DATE__", data.MaxDate)
	html = strings.ReplaceAll(html, "__INITIAL_FROM__", data.InitialFrom)
	html = strings.ReplaceAll(html, "__INITIAL_TO__", data.InitialTo)
	html = strings.ReplaceAll(html, "__GENERATED_AT__", data.GeneratedAt)
	html = strings.ReplaceAll(html, "__RECORDS_JSON__", data.RecordsJSON)
	f, err := os.CreateTemp("", "ai-telemetry-dashboard-*.html")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString(html); err != nil {
		return "", fmt.Errorf("write dashboard: %w", err)
	}
	return f.Name(), nil
}

// ── OpenBrowser ───────────────────────────────────────────────────────────────

func OpenBrowser(path string) error {
	url := "file://" + path
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func dayOnly(createdAt string) string {
	if len(createdAt) >= 10 {
		return createdAt[:10]
	}
	return ""
}

func resolveInitialRange(minDate, maxDate, rangeSpec string) (string, string, error) {
	if minDate == "" || maxDate == "" {
		return "", "", nil
	}

	spec := strings.TrimSpace(strings.ToLower(rangeSpec))
	if spec == "" {
		spec = "30d"
	}
	if spec == "all" {
		return minDate, maxDate, nil
	}

	days, ok := parseRangeDays(spec)
	if !ok || days <= 0 {
		days = 30
	}

	maxTime, err := time.Parse("2006-01-02", maxDate)
	if err != nil {
		return "", "", fmt.Errorf("parse maxDate %q: %w", maxDate, err)
	}
	fromTime := maxTime.AddDate(0, 0, -(days - 1))
	fromDate := fromTime.Format("2006-01-02")
	if fromDate < minDate {
		fromDate = minDate
	}
	return fromDate, maxDate, nil
}

func parseRangeDays(spec string) (int, bool) {
	if strings.HasSuffix(spec, "d") {
		spec = strings.TrimSuffix(spec, "d")
	}
	days, err := strconv.Atoi(spec)
	if err != nil {
		return 0, false
	}
	return days, true
}
