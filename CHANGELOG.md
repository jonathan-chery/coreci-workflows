# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] — 2026-05-26

### Added
- **`dag` package** — Directed Acyclic Graph engine
  - `dag.NewGraph(jobs)` — construct a graph from job definitions with dependency edges
  - `dag.Graph.TopoSort()` — Kahn's algorithm topological sort with deterministic ordering
  - `dag.Graph.Roots()` — return all nodes with zero dependencies
  - Cycle detection via DFS with recursive stack tracking; fails fast with descriptive errors
  - Implicit depth computation for execution stage scheduling
  - Unknown dependency validation — errors if a job references a non-existent dependency

- **`pipeline` package** — `.coreci.yml` parser and validator
  - `pipeline.Parse(data)` — unmarshal YAML into typed `Pipeline` struct
  - `pipeline.Load(path)` — read file and parse in one call
  - `pipeline.EvaluateRules(job, env)` — conditional job execution based on environment variables
  - Structural validation: duplicate job names, mutual-exclusivity of `plugin`/`invoke`, required fields
  - Full support for `jobs`, `services`, `env`, and `rules` YAML structures

- **Standard documentation** (`STANDARD.md`)
  - Complete YAML schema definition with field tables and constraint descriptions
  - Integration guide for external agents and orchestrators
  - Import path and version policy documentation
  - API stability contract (semver for schema, major bumps for breaking YAML changes)

- **CI / CD**
  - GitHub Actions workflow: test (Go 1.25.x), lint (`go vet` + `gofmt`), build
  - Coverage artifact upload on every PR

### Integration
```bash
go get github.com/jonathan-chery/coreci-workflows@v1.0.0
```

[Unreleased]: https://github.com/jonathan-chery/coreci-workflows/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/jonathan-chery/coreci-workflows/releases/tag/v1.0.0
