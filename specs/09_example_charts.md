# Example Charts

This document describes recommended MVP charts for `ai-log-report`.

## 1. Work Type Distribution

**Chart type:** bar chart or radar chart  
**Metric options:** task count, `SUM(estimated_time_min)`

**Use:** shows where AI effort is concentrated.

---

## 2. Complexity Distribution

**Chart type:** bar chart  
**Metric:** count of records

**Use:** shows whether the workload is mostly low, medium, or high complexity.

**Breakdowns worth supporting:**
- overall
- by model
- by task type

---

## 3. Interruptions by Model

**Chart type:** horizontal bar chart  
**Metric:** count of `task_type = interruption`

**Use:** compares how often different models or agents encounter blocked work.

---

## 4. Estimated Time by Model

**Chart type:** grouped bar chart  
**Metric:** `SUM(estimated_time_min)`

**Use:** shows which models account for the most AI-estimated effort.

---

## 5. Recommended vs Custom Tag Usage

**Chart type:** stacked bar or donut chart  
**Metric:** count of records by `*_tag_source`

**Use:** helps evaluate taxonomy drift and standardization quality.

---

## 6. Top Custom Tags

**Chart type:** ranked bar chart  
**Metric:** count from `task_tags`

**Use:** surfaces emerging categories that may deserve promotion to the starter vocabulary.

---

## 7. Average Confidence by Work Type

**Chart type:** bar chart  
**Metric:** `AVG(confidence)`

**Use:** shows which kinds of work agents classify most confidently.

---

## 8. Parent Link Status

**Chart type:** donut chart or bar chart  
**Metric:** count by `parent_link_status`

**Use:** reveals how often subtasks or interruptions are linked, dangling, or standalone.

---

## 9. Model × Complexity Heatmap

**Chart type:** heatmap  
**Metric:** `SUM(estimated_time_min)` or count

**Use:** quickly shows how workload shape differs by model.

---

## 10. Work Type × Complexity Heatmap

**Chart type:** heatmap  
**Metric:** `SUM(estimated_time_min)`

**Use:** shows what kinds of work are being estimated as hardest.

---

# Suggested MVP Dashboard Layout

## Section 1 — Overview
- total tasks
- total subtasks
- total interruptions
- total estimated minutes

## Section 2 — Core Distribution
- work type distribution
- complexity distribution
- interruptions by model

## Section 3 — Taxonomy Quality
- recommended vs custom tags
- top custom tags
- null language/domain counts

## Section 4 — Deep Dive
- average confidence by work type
- model × complexity heatmap
- parent link status
