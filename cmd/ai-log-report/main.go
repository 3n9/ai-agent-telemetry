package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/3n9/ai-agent-telemetry/internal/dashboard"
	"github.com/3n9/ai-agent-telemetry/internal/db"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:               "ai-log-report",
		Short:             "AI agent telemetry reporting tool",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	root.AddCommand(summaryCmd(), chartCmd(), exportCmd(), dashboardCmd(), logsCmd())
	return root
}

// ── helpers ───────────────────────────────────────────────────────────────────

func openDB() (*sql.DB, error) {
	path, err := db.DBPath()
	if err != nil {
		return nil, err
	}
	conn, err := db.Open(path)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ── summary ───────────────────────────────────────────────────────────────────

func summaryCmd() *cobra.Command {
	var by string
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show telemetry statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := openDB()
			if err != nil {
				return err
			}
			defer conn.Close()

			if by == "" {
				return runOverallSummary(conn)
			}
			return runSummaryBy(conn, by)
		},
	}
	cmd.Flags().StringVar(&by, "by", "", "Group by dimension: work_type, model_name, complexity, domain, language, task_type")
	return cmd
}

func runOverallSummary(conn *sql.DB) error {
	rows, err := conn.Query(`
		SELECT task_type, COUNT(*) FROM tasks GROUP BY task_type
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var tt string
		var n int
		if err := rows.Scan(&tt, &n); err != nil {
			return err
		}
		counts[tt] = n
	}

	var totalMin int
	_ = conn.QueryRow(`SELECT COALESCE(SUM(estimated_time_min),0) FROM tasks`).Scan(&totalMin)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Metric\tCount")
	fmt.Fprintln(w, "------\t-----")
	fmt.Fprintf(w, "Total tasks\t%d\n", counts["task"])
	fmt.Fprintf(w, "Subtasks\t%d\n", counts["subtask"])
	fmt.Fprintf(w, "Interruptions\t%d\n", counts["interruption"])
	fmt.Fprintf(w, "Total estimated minutes\t%d\n", totalMin)
	return w.Flush()
}

var allowedDimensions = map[string]bool{
	"work_type": true, "model_name": true, "complexity": true,
	"domain": true, "language": true, "task_type": true,
}

func runSummaryBy(conn *sql.DB, dimension string) error {
	if !allowedDimensions[dimension] {
		return fmt.Errorf("unknown dimension %q — choose: work_type, model_name, complexity, domain, language, task_type", dimension)
	}
	query := fmt.Sprintf(`
		SELECT COALESCE(%s,'(null)'), COUNT(*), COALESCE(SUM(estimated_time_min),0)
		FROM tasks GROUP BY %s ORDER BY COUNT(*) DESC
	`, dimension, dimension)
	rows, err := conn.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\tCount\tEst. minutes\n", strings.ToUpper(dimension))
	fmt.Fprintln(w, strings.Repeat("-", 20)+"\t-----\t------------")
	for rows.Next() {
		var dim string
		var count, mins int
		if err := rows.Scan(&dim, &count, &mins); err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\t%d\t%d\n", dim, count, mins)
	}
	return w.Flush()
}

// ── chart ─────────────────────────────────────────────────────────────────────

func chartCmd() *cobra.Command {
	var metric string
	var by string
	cmd := &cobra.Command{
		Use:   "chart [bar|pie|radar]",
		Short: "Display a terminal chart",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := openDB()
			if err != nil {
				return err
			}
			defer conn.Close()

			chartType := args[0]
			if metric == "" {
				metric = "count"
			}
			if by == "" {
				by = "work_type"
			}
			if !allowedDimensions[by] {
				return fmt.Errorf("unknown dimension %q — choose: work_type, model_name, complexity, domain, language, task_type", by)
			}

			data, labels, err := fetchChartData(conn, metric, by)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				fmt.Println("No data available.")
				return nil
			}

			switch chartType {
			case "bar", "radar":
				return renderBar(labels, data, metric, by)
			case "pie":
				return renderPie(labels, data, by)
			default:
				return fmt.Errorf("unsupported chart type %q — choose: bar, pie, radar", chartType)
			}
		},
	}
	cmd.Flags().StringVar(&metric, "metric", "count", "Metric: count or estimated_time_min")
	cmd.Flags().StringVar(&by, "by", "work_type", "Group by dimension: work_type, model_name, complexity, domain, language, task_type")
	return cmd
}

func fetchChartData(conn *sql.DB, metric, dimension string) ([]float64, []string, error) {
	var selectExpr string
	switch metric {
	case "estimated_time_min":
		selectExpr = "COALESCE(SUM(estimated_time_min),0)"
	default:
		selectExpr = "COUNT(*)"
	}
	query := fmt.Sprintf(
		`SELECT COALESCE(%s,'(null)'), %s FROM tasks GROUP BY %s ORDER BY %s DESC`,
		dimension, selectExpr, dimension, selectExpr,
	)
	rows, err := conn.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var labels []string
	var values []float64
	for rows.Next() {
		var label string
		var val float64
		if err := rows.Scan(&label, &val); err != nil {
			return nil, nil, err
		}
		labels = append(labels, label)
		values = append(values, val)
	}
	return values, labels, nil
}

const barWidth = 40

func renderBar(labels []string, values []float64, metric, dimension string) error {
	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	maxLabel := 0
	for _, l := range labels {
		if len(l) > maxLabel {
			maxLabel = len(l)
		}
	}

	fmt.Printf("\n%s Distribution (%s)\n\n", strings.ToUpper(dimension), metric)
	for i, label := range labels {
		bar := int(math.Round(values[i] / maxVal * barWidth))
		fmt.Printf("%-*s │ %s %.0f\n", maxLabel, label, strings.Repeat("█", bar), values[i])
	}
	fmt.Println()
	return nil
}

func renderPie(labels []string, values []float64, dimension string) error {
	total := 0.0
	for _, v := range values {
		total += v
	}
	fmt.Println()
	fmt.Printf("%s Distribution (pie)\n", strings.ToUpper(dimension))
	fmt.Println()
	for i, label := range labels {
		pct := values[i] / total * 100
		bar := int(math.Round(pct / 100 * barWidth))
		fmt.Printf("%-16s %5.1f%% %s\n", label, pct, strings.Repeat("▓", bar))
	}
	fmt.Println()
	return nil
}

// ── dashboard ─────────────────────────────────────────────────────────────────

func dashboardCmd() *cobra.Command {
	var noOpen bool
	var dateRange string
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Generate and open an HTML dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := openDB()
			if err != nil {
				return err
			}
			defer conn.Close()

			data, err := dashboard.Build(conn, dashboard.BuildOptions{
				InitialRange: dateRange,
			})
			if err != nil {
				return fmt.Errorf("building dashboard data: %w", err)
			}

			path, err := dashboard.Render(data)
			if err != nil {
				return fmt.Errorf("rendering dashboard: %w", err)
			}

			fmt.Fprintf(os.Stderr, "dashboard: %s\n", path)

			if !noOpen {
				if err := dashboard.OpenBrowser(path); err != nil {
					fmt.Fprintf(os.Stderr, "could not open browser: %v\n", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Write HTML file without opening browser")
	cmd.Flags().StringVar(&dateRange, "range", "30d", "Initial dashboard date range: Nd or all")
	return cmd
}

// ── logs ──────────────────────────────────────────────────────────────────────

func logsCmd() *cobra.Command {
	var limit int
	var agent string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show raw log records (for debugging)",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := openDB()
			if err != nil {
				return err
			}
			defer conn.Close()

			// Build WHERE clause
			where := "1=1"
			var qargs []any
			if agent != "" {
				where += " AND agent_name = ?"
				qargs = append(qargs, agent)
			}

			query := fmt.Sprintf(`
				SELECT t.id, t.created_at, t.agent_name, t.model_name,
				       t.work_type, COALESCE(t.secondary_work_type,''),
				       t.complexity, t.confidence, t.estimated_time_min,
				       t.task_type, COALESCE(t.parent_task_id,''),
				       COALESCE(GROUP_CONCAT(tt.tag_value, ' '), '') AS tags
				FROM tasks t
				LEFT JOIN task_tags tt ON tt.task_id = t.id
				WHERE %s
				GROUP BY t.id
				ORDER BY t.created_at DESC
				LIMIT ?`, where)
			qargs = append(qargs, limit)

			rows, err := conn.Query(query, qargs...)
			if err != nil {
				return err
			}
			defer rows.Close()

			type row struct {
				ID           string  `json:"id"`
				CreatedAt    string  `json:"created_at"`
				Agent        string  `json:"agent_name"`
				Model        string  `json:"model_name"`
				WorkType     string  `json:"work_type"`
				SecWorkType  string  `json:"secondary_work_type,omitempty"`
				Complexity   string  `json:"complexity"`
				Confidence   float64 `json:"confidence"`
				EstMins      int     `json:"estimated_time_min"`
				TaskType     string  `json:"task_type"`
				ParentTaskID string  `json:"parent_task_id,omitempty"`
				Tags         string  `json:"tags,omitempty"`
			}

			var records []row
			for rows.Next() {
				var r row
				if err := rows.Scan(&r.ID, &r.CreatedAt, &r.Agent, &r.Model,
					&r.WorkType, &r.SecWorkType, &r.Complexity, &r.Confidence,
					&r.EstMins, &r.TaskType, &r.ParentTaskID, &r.Tags); err != nil {
					return err
				}
				records = append(records, r)
			}

			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(records)
			}

			if len(records) == 0 {
				fmt.Println("no records found")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "CREATED AT\tID\tAGENT\tMODEL\tWORK TYPE\tCOMPLEXITY\tCONF\tMINS\tTYPE\tTAGS")
			fmt.Fprintln(w, strings.Repeat("─", 10)+"\t"+strings.Repeat("─", 26)+"\t"+
				strings.Repeat("─", 12)+"\t"+strings.Repeat("─", 20)+"\t"+
				strings.Repeat("─", 14)+"\t"+strings.Repeat("─", 10)+"\t"+
				strings.Repeat("─", 4)+"\t"+strings.Repeat("─", 4)+"\t"+
				strings.Repeat("─", 13)+"\t"+strings.Repeat("─", 20))
			for _, r := range records {
				ts := r.CreatedAt
				if len(ts) >= 16 {
					ts = ts[:10] + " " + ts[11:16]
				}
				wt := r.WorkType
				if r.SecWorkType != "" {
					wt += "/" + r.SecWorkType
				}
				parent := ""
				if r.ParentTaskID != "" {
					parent = " ↳" + r.ParentTaskID[:8]
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%.2f\t%d\t%s%s\t%s\n",
					ts, r.ID, r.Agent, r.Model, wt,
					r.Complexity, r.Confidence, r.EstMins, r.TaskType, parent, r.Tags)
			}
			return w.Flush()
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of records to show (most recent first)")
	cmd.Flags().StringVar(&agent, "agent", "", "Filter by agent name")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

// ── export ────────────────────────────────────────────────────────────────────
func exportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export [csv|json]",
		Short: "Export telemetry data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := openDB()
			if err != nil {
				return err
			}
			defer conn.Close()

			switch args[0] {
			case "csv":
				return exportCSV(conn)
			case "json":
				return exportJSON(conn)
			default:
				return fmt.Errorf("unknown format %q — choose: csv, json", args[0])
			}
		},
	}
}

const exportQuery = `
SELECT id, created_at, schema_version, agent_name, model_name,
       work_type, work_type_tag_source,
       COALESCE(secondary_work_type,''), COALESCE(secondary_work_type_tag_source,''),
       COALESCE(language,''), COALESCE(language_tag_source,''),
       COALESCE(domain,''), COALESCE(domain_tag_source,''),
       complexity, confidence, estimated_time_min,
       task_type, COALESCE(parent_task_id,''),
       COALESCE(input_tokens,-1), COALESCE(output_tokens,-1), COALESCE(cost_estimate,-1)
FROM tasks ORDER BY created_at
`

var csvHeaders = []string{
	"id", "created_at", "schema_version", "agent_name", "model_name",
	"work_type", "work_type_tag_source",
	"secondary_work_type", "secondary_work_type_tag_source",
	"language", "language_tag_source",
	"domain", "domain_tag_source",
	"complexity", "confidence", "estimated_time_min",
	"task_type", "parent_task_id",
	"input_tokens", "output_tokens", "cost_estimate",
}

func exportCSV(conn *sql.DB) error {
	rows, err := conn.Query(exportQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	w := csv.NewWriter(os.Stdout)
	_ = w.Write(csvHeaders)
	for rows.Next() {
		rec := make([]any, len(csvHeaders))
		ptrs := make([]any, len(csvHeaders))
		for i := range rec {
			ptrs[i] = &rec[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		row := make([]string, len(rec))
		for i, v := range rec {
			row[i] = fmt.Sprintf("%v", v)
		}
		_ = w.Write(row)
	}
	w.Flush()
	return w.Error()
}

func exportJSON(conn *sql.DB) error {
	rows, err := conn.Query(`SELECT raw_payload_json, id, created_at FROM tasks ORDER BY created_at`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type record struct {
		ID        string          `json:"id"`
		CreatedAt string          `json:"created_at"`
		Payload   json.RawMessage `json:"payload"`
	}

	var records []record
	for rows.Next() {
		var raw, id, createdAt string
		if err := rows.Scan(&raw, &id, &createdAt); err != nil {
			return err
		}
		records = append(records, record{ID: id, CreatedAt: createdAt, Payload: json.RawMessage(raw)})
	}

	sort.Slice(records, func(i, j int) bool { return records[i].CreatedAt < records[j].CreatedAt })

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}
