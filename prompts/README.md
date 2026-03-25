# AI Agent Telemetry — Prompts

This directory contains ready-to-use system prompts that instruct AI agents to log their activity using the `ai-log` tool.

## Global Installation (Recommended)

**Standalone (no repo clone required):**

```sh
curl -fsSL https://raw.githubusercontent.com/3n9/ai-agent-telemetry/main/scripts/install-global.sh | sh
```

**From a repo clone:**

```sh
make install-global
```

This will:
- Sync prompts to `~/.ai-telemetry/prompts/`
- Configure a **BeforeAgent hook** for Gemini CLI in `~/.gemini/settings.json`
- Append instructions to `~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md`
- Create a specific instruction file for Copilot CLI at `~/.copilot/ai-telemetry.instructions.md`
- Set up global conventions for Aider at `~/.aider.conventions.md`

## Per-Project Usage

| File | Agent | How to apply |
|---|---|---|
| `system-prompt.md` | Any agent | Copy into your system prompt |
| `claude-code.md` | Claude Code | Add to `CLAUDE.md` in your project root |
| `copilot.md` | GitHub Copilot CLI | Add to `.github/copilot-instructions.md` |
| `gemini.md` | Gemini CLI | Add to `GEMINI.md` in your project root |
| `codex.md` | OpenAI Codex CLI | Add to `AGENTS.md` in your project root |

## Prerequisites

`ai-log` must be installed and on `$PATH`. Run once to initialise the database:

```sh
ai-log init
```

## Payload quick reference

```json
{
  "schema_version": 1,
  "agent_name": "<your-agent-name>",
  "model_name": "<model-in-use>",
  "work_type": "coding",
  "complexity": "medium",
  "confidence": 0.85,
  "estimated_time_min": 15,
  "task_type": "task"
}
```

Full field reference: see `../specs/05_technical_spec.md`
