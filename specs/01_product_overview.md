# AI Agent Telemetry Logger — Product Overview

## Summary
AI Agent Telemetry Logger is a lightweight telemetry system designed to capture **self-reported activity from AI CLI agents**.
Agents log structured metadata describing the work they are performing. The data can later be analyzed to understand how AI tools are used.

The system focuses on **semantic telemetry**, not runtime monitoring. Instead of observing AI activity externally, agents declare their work.

The product consists of two tools:

- **ai-log** — telemetry ingestion tool
- **ai-log-report** — reporting and analytics tool

These tools together allow analysis of AI workflows while keeping the logging path lightweight.

---

## Product Goals

The system helps answer questions such as:

- What types of work AI spends time on
- Which models are used most frequently
- How complex tasks are
- How often interruptions occur
- What kinds of work agents estimate as difficult

Example insights:

- distribution of work types
- complexity distribution
- interruption rate by model
- most common custom tags
- average estimated effort per task type

---

## Core Philosophy

### Self-reported telemetry
AI agents declare their own activities.

### Lightweight logging
Logging must be extremely simple so agents can call it frequently.

### Structured metadata
All telemetry follows a consistent schema so analytics remain meaningful.

### Open taxonomy
Agents may invent tags, allowing categories to evolve organically.

### Privacy first
The system intentionally avoids logging sensitive data:

- project names
- file paths
- source code
- prompts
- user messages

Only **task metadata** is recorded.

---

## Task Model

### Task
Primary declared unit of work.

Examples:
- implementing a feature
- writing documentation
- researching a topic

### Subtask
A meaningful unit of work within another task.

Subtasks may reference parent tasks that have not yet been logged.

### Interruption
Represents blocked or abandoned work.

Examples:
- missing context
- dependency failure
- tool malfunction
- abandoned approach

Interruptions may optionally reference a parent task.

---

## Starter Vocabulary

### Suggested work types

- coding
- debugging
- research
- analysis
- writing
- planning
- creative
- support
- refactor

These are **recommended**, not mandatory.

### Suggested languages

- php
- javascript
- typescript
- python
- sql
- html
- css
- shell
- json
- yaml
- markdown
- none

### Suggested domains

- frontend
- backend
- database
- devops
- documentation
- wordpress
- laravel
- api
- testing
- fiction
- horror
- email
- blog
- marketing
- none
