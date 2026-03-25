package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/3n9/ai-agent-telemetry/internal/db"
	"github.com/3n9/ai-agent-telemetry/service"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:               "ai-log",
		Short:             "AI agent telemetry ingestion tool",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	root.AddCommand(initCmd(), emitCmd(), validateCmd(), resetCmd())
	return root
}

// ── init ─────────────────────────────────────────────────────────────────────

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create the telemetry database",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := db.DBPath()
			if err != nil {
				return writeError("DB_PATH_ERROR", err.Error())
			}
			conn, err := db.Open(path)
			if err != nil {
				return writeError("DB_OPEN_ERROR", err.Error())
			}
			defer conn.Close()
			if err := db.Init(conn); err != nil {
				return writeError("DB_INIT_ERROR", err.Error())
			}
			fmt.Fprintf(os.Stderr, "database initialised: %s\n", path)
			return nil
		},
	}
}

// ── validate ─────────────────────────────────────────────────────────────────

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate '<json>'",
		Short: "Validate a payload without storing it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := service.ParseAndValidate(args[0])
			if err != nil {
				return writeError(err.Code, err.Message)
			}
			return writeJSON(map[string]any{
				"ok":             true,
				"task_type":      p.TaskType,
				"schema_version": p.SchemaVersion,
				"warnings":       []any{},
			})
		},
	}
}

// ── emit ──────────────────────────────────────────────────────────────────────

func emitCmd() *cobra.Command {
	var parentTaskID string
	var taskID string

	cmd := &cobra.Command{
		Use:   "emit '<json>'",
		Short: "Store a telemetry record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := service.EmitRequest{RawPayloadJSON: args[0]}
			if cmd.Flags().Changed("task-id") {
				req.TaskIDOverride = &taskID
			}
			if cmd.Flags().Changed("parent-task-id") {
				req.ParentTaskOverride = &parentTaskID
			}

			resp, err := service.Emit(req)
			if err != nil {
				return writeError(err.Code, err.Message)
			}
			return writeJSON(resp)
		},
	}

	cmd.Flags().StringVar(&taskID, "task-id", "", "Override task ID")
	cmd.Flags().StringVar(&parentTaskID, "parent-task-id", "", "Set parent task ID")
	return cmd
}

// ── reset ─────────────────────────────────────────────────────────────────────

func resetCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Empty the telemetry database (keeps schema)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Fprint(os.Stderr, "This will delete all telemetry records. Use --force to confirm.\n")
				os.Exit(1)
			}
			path, err := db.DBPath()
			if err != nil {
				return writeError("DB_PATH_ERROR", err.Error())
			}
			conn, err := db.Open(path)
			if err != nil {
				return writeError("DB_OPEN_ERROR", err.Error())
			}
			defer conn.Close()
			if err := db.Reset(conn); err != nil {
				return writeError("DB_RESET_ERROR", err.Error())
			}
			fmt.Fprintln(os.Stderr, "database reset: all records deleted")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion of all records")
	return cmd
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeError(code, message string) error {
	_ = writeJSON(map[string]any{
		"ok":    false,
		"error": map[string]string{"code": code, "message": message},
	})
	os.Exit(1)
	return nil
}
