# ai-log-report Specification

## Overview

`ai-log-report` is the reporting and analytics CLI for telemetry stored by ai-log.

The tool is read-only and never modifies stored telemetry.

---

## Commands

### Summary

ai-log-report summary

Displays high level statistics.

Example output:

Total tasks: 452
Subtasks: 211
Interruptions: 37

---

### Summary by dimension

ai-log-report summary --by work_type
ai-log-report summary --by model_name
ai-log-report summary --by complexity

---

### Charts

ai-log-report chart radar --metric estimated_time_min

Supported chart types:

- radar
- bar
- pie
- treemap
- heatmap

---

### Dashboard

ai-log-report dashboard

Launches a local web dashboard.

Potential visualization libraries:

- ECharts
- Plotly
- Vega

---

### Export

ai-log-report export csv
ai-log-report export json

Exports telemetry data for external analysis.

---

## Default Reports

Recommended built-in reports:

- Work distribution
- Complexity distribution
- Interruptions by model
- Top custom tags
- Average confidence per work type
- Recommended vs custom tag usage
