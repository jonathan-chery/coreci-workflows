# CoreCI Workflow Standard

## Overview

This document defines the **CoreCI Workflow Standard** — the YAML schema, semantics, and API contract for pipeline definitions used across all CoreCI-related projects.

## Schema Definition

A valid `.coreci.yml` file is a single YAML document. The top-level keys are `jobs`, `services`, and `env`.

### Top-Level Structure

```yaml
jobs:     # Required — map[string]Job
services: # Optional — map[string]Service
env:      # Optional — EnvBlock
```

### Job

A job is a named execution unit within a pipeline.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `needs` | `[]string` | no | Job IDs that must complete before this job runs |
| `plugin` | `string` | <1 of plugin/invoke> | Execution plugin (e.g. `docker://image:tag`) |
| `invoke` | `string` | <1 of plugin/invoke> | Inline shell command |
| `vars` | `map[string]string` | no | Job-level environment variables |
| `memory_limit_mb` | `int` | no | Memory limit in megabytes |
| `timeout_ms` | `int` | no | Timeout in milliseconds |
| `rules` | `[]Rule` | no | Conditional execution rules |
| `tags` | `[]string` | no | Runner tags for routing |

**Constraint:** Exactly one of `plugin` or `invoke` must be set. Both or neither is invalid.

**Constraint:** Duplicate job names are forbidden.

```yaml
jobs:
  build:
    plugin: docker://golang:1.22
    vars:
      CGO_ENABLED: "0"
  test:
    needs: [build]
    plugin: docker://golang:1.22
    memory_limit_mb: 512
    timeout_ms: 60000
    tags: [gpu]
```

### Service

A service describes an external dependency available during pipeline execution.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | yes | Service identifier |
| `image` | `string` | yes | Container image |
| `plugin` | `string` | no | Custom plugin for provisioning |
| `engine` | `string` | no | Container engine override |
| `acl` | `[]string` | no | Access control list |
| `ports` | `[]string` | no | Port mappings (e.g. `"6379:6379"`) |
| `vars` | `map[string]string` | no | Service-specific env vars |
| `persistent` | `bool` | no | Whether to keep the service after the run |

```yaml
services:
  redis:
    image: redis:7
    ports:
      - "6379:6379"
  postgres:
    image: postgres:16
    persistent: true
```

### EnvBlock

Controls environment variable injection.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `from_ci` | `[]string` | no | CI variable names to auto-inject |
| `vars` | `map[string]string` | no | Static key-value pairs |

```yaml
env:
  from_ci:
    - CI_COMMIT_SHA
    - CI_COMMIT_BRANCH
  vars:
    FOO: bar
```

### Rule

Controls conditional job execution.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `if` | `string` | no | Environment variable name; truthy if non-empty |
| `when` | `string` | no | `"on_success"` (default), `"always"`, or `"never"` |

**Rule semantics:**
- If no rules are defined, the job always runs.
- Each rule is evaluated in order. The first matching rule controls inclusion.
- A rule matches if `if` is empty (unconditional) or the named env var is non-empty.
- If `when: never`, the rule excludes the job regardless of match.
- If no rule matches, the job is excluded.

```yaml
rules:
  - if: "CI_COMMIT_TAG"      # run if tag exists
  - if: "CI_COMMIT_BRANCH"
    when: "never"             # but skip if branch exists
```

## Integration Guide

### As a Pipeline Author

1. Write a `.coreci.yml` file following the schema above.
2. Use `needs` to express dependencies between jobs.
3. Use `rules` to conditionally skip jobs (e.g. skip deploy on PRs).

### As a Backlog Agent / Orchestrator

Import the standard library to parse and validate pipeline files:

```go
import (
    "github.com/jonathan-chery/coreci-workflows/pipeline"
    "github.com/jonathan-chery/coreci-workflows/dag"
)

func executePipeline(rawYAML []byte) error {
    p, err := pipeline.Parse(rawYAML)
    if err != nil {
        return fmt.Errorf("invalid pipeline: %w", err)
    }

    rawJobs := make(map[string]struct{ Needs []string }, len(p.Jobs))
    for id, j := range p.Jobs {
        rawJobs[id] = struct{ Needs []string }{Needs: j.Needs}
    }

    g, err := dag.NewGraph(rawJobs)
    if err != nil {
        return fmt.Errorf("dag: %w", err)
    }

    for _, id := range g.TopoSort() {
        job := p.Jobs[id]
        if !pipeline.EvaluateRules(job, env) {
            continue
        }
        // execute job ...
    }
    return nil
}
```

### Import Instructions

**Module:** `github.com/jonathan-chery/coreci-workflows`

```bash
go get github.com/jonathan-chery/coreci-workflows@v1.0.0
```

**Packages:**
- `github.com/jonathan-chery/coreci-workflows/dag` — Graph operations
- `github.com/jonathan-chery/coreci-workflows/pipeline` — YAML parsing and validation

## API Stability

This standard is versioned with semver. Minor version bumps may add fields. Patch version bumps are bug fixes only. Major version bumps indicate breaking changes to the YAML schema.

**Current version:** v1.0.0

## Changelog

### v1.0.0 (2026-05-26)
- Initial extraction from CoreCI v1.1
- `dag`: Graph construction, topological sort, cycle detection, depth computation
- `pipeline`: YAML parsing, structural validation, rule evaluation
- Supports `jobs`, `services`, `env`, and `rules` structures
