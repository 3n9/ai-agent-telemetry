package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	reportBinary string
	logBinary    string
)

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "ai-log-report-test-*")
	if err != nil {
		os.Exit(1)
	}

	reportBinary = filepath.Join(tmp, "ai-log-report")
	cmd1 := exec.Command("go", "build", "-o", reportBinary, ".")
	cmd1.Dir = "."
	if out, err := cmd1.CombinedOutput(); err != nil {
		_, _ = os.Stderr.Write(out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	logBinary = filepath.Join(tmp, "ai-log")
	cmd2 := exec.Command("go", "build", "-o", logBinary, "github.com/3n9/ai-agent-telemetry/cmd/ai-log")
	if out, err := cmd2.CombinedOutput(); err != nil {
		_, _ = os.Stderr.Write(out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

const validPayload = `{"schema_version":1,"agent_name":"test-agent","model_name":"test-model","work_type":"coding","complexity":"low","confidence":0.9,"estimated_time_min":5,"task_type":"task"}`

// runBin executes a binary with the given args, inheriting the current test env.
func runBin(t *testing.T, binary string, args ...string) (stdout string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Env = os.Environ()
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return outBuf.String(), exitErr.ExitCode()
		}
		return outBuf.String(), 1
	}
	return outBuf.String(), 0
}

// seedDB initialises the schema and emits one record via the ai-log binary.
func seedDB(t *testing.T) {
	t.Helper()
	if _, code := runBin(t, logBinary, "init"); code != 0 {
		t.Fatal("ai-log init failed")
	}
	if _, code := runBin(t, logBinary, "emit", validPayload); code != 0 {
		t.Fatal("ai-log emit failed")
	}
}

// initDB only initialises the schema (no records).
func initDB(t *testing.T) {
	t.Helper()
	if _, code := runBin(t, logBinary, "init"); code != 0 {
		t.Fatal("ai-log init failed")
	}
}

func TestSummaryEmpty(t *testing.T) {
	t.Setenv("AI_LOG_DB", filepath.Join(t.TempDir(), "test.db"))
	initDB(t)

	out, code := runBin(t, reportBinary, "summary")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stdout: %s", code, out)
	}
	if !strings.Contains(out, "Total tasks") {
		t.Fatalf("expected 'Total tasks' header in output, got:\n%s", out)
	}
}

func TestSummaryWithData(t *testing.T) {
	t.Setenv("AI_LOG_DB", filepath.Join(t.TempDir(), "test.db"))
	seedDB(t)

	out, code := runBin(t, reportBinary, "summary")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stdout: %s", code, out)
	}
	// runOverallSummary outputs a row "Total tasks  <n>"; confirm count is 1
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Total tasks") && strings.Contains(line, "1") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'Total tasks' count of 1 in summary:\n%s", out)
	}
}

func TestSummaryByDimension(t *testing.T) {
	t.Setenv("AI_LOG_DB", filepath.Join(t.TempDir(), "test.db"))
	seedDB(t)

	out, code := runBin(t, reportBinary, "summary", "--by", "model_name")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stdout: %s", code, out)
	}
	if !strings.Contains(strings.ToUpper(out), "MODEL_NAME") {
		t.Fatalf("expected MODEL_NAME header in output, got:\n%s", out)
	}
}

func TestChartBar(t *testing.T) {
	t.Setenv("AI_LOG_DB", filepath.Join(t.TempDir(), "test.db"))
	seedDB(t)

	_, code := runBin(t, reportBinary, "chart", "bar")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestChartByDimension(t *testing.T) {
	t.Setenv("AI_LOG_DB", filepath.Join(t.TempDir(), "test.db"))
	seedDB(t)

	_, code := runBin(t, reportBinary, "chart", "bar", "--by", "complexity")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestLogsJSON(t *testing.T) {
	t.Setenv("AI_LOG_DB", filepath.Join(t.TempDir(), "test.db"))
	seedDB(t)

	out, code := runBin(t, reportBinary, "logs", "--json")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stdout: %s", code, out)
	}
	var records []map[string]any
	if err := json.Unmarshal([]byte(out), &records); err != nil {
		t.Fatalf("expected valid JSON array: %v\nstdout: %s", err, out)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}

func TestExportCSV(t *testing.T) {
	t.Setenv("AI_LOG_DB", filepath.Join(t.TempDir(), "test.db"))
	initDB(t)

	out, code := runBin(t, reportBinary, "export", "csv")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stdout: %s", code, out)
	}
	header := strings.SplitN(strings.TrimSpace(out), "\n", 2)[0]
	if !strings.Contains(header, "id") || !strings.Contains(header, "agent_name") {
		t.Fatalf("expected CSV header with 'id' and 'agent_name', got: %s", header)
	}
}

func TestExportJSON(t *testing.T) {
	t.Setenv("AI_LOG_DB", filepath.Join(t.TempDir(), "test.db"))
	seedDB(t)

	out, code := runBin(t, reportBinary, "export", "json")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stdout: %s", code, out)
	}
	var records []map[string]any
	if err := json.Unmarshal([]byte(out), &records); err != nil {
		t.Fatalf("expected valid JSON array: %v\nstdout: %s", err, out)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}
