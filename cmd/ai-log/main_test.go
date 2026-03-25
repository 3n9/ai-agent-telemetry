package main_test

import (
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "ai-log-test-*")
	if err != nil {
		os.Exit(1)
	}
	binaryPath = filepath.Join(tmp, "ai-log")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		_, _ = os.Stderr.Write(out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

const validPayload = `{"schema_version":1,"agent_name":"test-agent","model_name":"test-model","work_type":"coding","complexity":"low","confidence":0.9,"estimated_time_min":5,"task_type":"task"}`

// run executes the ai-log binary with the given args, inheriting the test env
// (including any AI_LOG_DB set via t.Setenv). Returns stdout and the exit code.
func run(t *testing.T, args ...string) (stdout string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
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

func TestInitCreatesDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("AI_LOG_DB", dbPath)

	_, code := run(t, "init")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("expected DB file to exist after init")
	}
}

func TestEmitWritesRecord(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("AI_LOG_DB", dbPath)

	if _, code := run(t, "init"); code != 0 {
		t.Fatal("init failed")
	}
	out, code := run(t, "emit", validPayload)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stdout: %s", code, out)
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nstdout: %s", err, out)
	}
	if ok, _ := resp["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got: %v", resp)
	}
	if taskID, _ := resp["task_id"].(string); taskID == "" {
		t.Fatalf("expected non-empty task_id, got: %v", resp)
	}
}

func TestEmitInvalidJSON(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("AI_LOG_DB", dbPath)

	out, code := run(t, "emit", "not-json")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	var resp map[string]any
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("expected error JSON in stdout, got: %s", out)
	}
	if ok, _ := resp["ok"].(bool); ok {
		t.Fatal("expected ok=false")
	}
}

func TestEmitMissingRequiredField(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("AI_LOG_DB", dbPath)

	// "complexity" is omitted — will be "" and fail the allowed-values check
	payload := `{"schema_version":1,"agent_name":"test-agent","model_name":"test-model","work_type":"coding","confidence":0.9,"estimated_time_min":5,"task_type":"task"}`
	out, code := run(t, "emit", payload)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d; stdout: %s", code, out)
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("expected error JSON, got: %s", out)
	}
	errObj, _ := resp["error"].(map[string]any)
	if errObj == nil {
		t.Fatalf("expected error object in response, got: %v", resp)
	}
	if errCode, _ := errObj["code"].(string); errCode != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got: %q", errCode)
	}
}

func TestValidateDryRun(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("AI_LOG_DB", dbPath)

	out, code := run(t, "validate", validPayload)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stdout: %s", code, out)
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, out)
	}
	if ok, _ := resp["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got: %v", resp)
	}
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Fatal("expected no DB file to be created by validate")
	}
}

func TestResetForce(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("AI_LOG_DB", dbPath)

	if _, code := run(t, "init"); code != 0 {
		t.Fatal("init failed")
	}
	if _, code := run(t, "emit", validPayload); code != 0 {
		t.Fatal("emit failed")
	}

	if _, code := run(t, "reset", "--force"); code != 0 {
		t.Fatal("reset --force failed")
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db after reset: %v", err)
	}
	defer conn.Close()

	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count); err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 records after reset, got %d", count)
	}
}
