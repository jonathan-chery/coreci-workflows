# CoreCI Workflows Domain

The **CoreCI Workflows** domain is the shared library that defines the pipeline standard for the entire workspace. It is the single source of truth for:

- Pipeline YAML parsing and validation
- DAG construction, topological sorting, and cycle detection
- Workflow rule evaluation

## Packages

| Package | Description |
|---------|-------------|
| `dag` | Directed Acyclic Graph engine — construction, topological sort, cycle detection, depth computation |
| `pipeline` | YAML parser for `.coreci.yml` — schema types, validation, rule evaluation |

## Installation

```bash
go get github.com/coreci/workflows@v1.0.0
```

## Usage

### Parse a pipeline file

```go
import "github.com/coreci/workflows/pipeline"

p, err := pipeline.Load(".coreci.yml")
if err != nil {
    log.Fatal(err)
}
for name, job := range p.Jobs {
    fmt.Println(name, job.Plugin)
}
```

### Build a DAG from parsed jobs

```go
import "github.com/coreci/workflows/dag"

jobs := map[string]struct{ Needs []string }{
    "build": {},
    "test":  {Needs: []string{"build"}},
}
g, err := dag.NewGraph(jobs)
if err != nil {
    log.Fatal(err)
}
order := g.TopoSort() // ["build", "test"]
```

### Evaluate job rules

```go
env := map[string]string{"CI_COMMIT_TAG": "v1.0.0"}
if pipeline.EvaluateRules(job, env) {
    // job should run
}
```

## Version

Current: **v1.0.0**

## License

MIT
