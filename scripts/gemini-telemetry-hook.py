#!/usr/bin/env python3
import sys
import json
import os

def run_hook():
    # Read the prompt file
    prompt_path = os.path.expanduser("~/.ai-telemetry/prompts/gemini.md")
    
    try:
        with open(prompt_path, "r") as f:
            prompt_content = f.read()
    except Exception as e:
        # Fallback if file is missing
        print(json.dumps({"hookSpecificOutput": {}}))
        return

    # Wrap in the required JSON structure for Gemini CLI
    output = {
        "hookSpecificOutput": {
            "additionalContext": f"\n\n### AI TELEMETRY POLICY\n{prompt_content}"
        }
    }
    
    print(json.dumps(output))

if __name__ == "__main__":
    # Consume stdin to avoid pipe issues
    if not sys.stdin.isatty():
        sys.stdin.read()
    run_hook()
