// Package pipeline parses the .coreci.yml pipeline definition into domain models.
package pipeline

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// EnvBlock controls how environment variables and CI context are injected into
// the pipeline. If FromCI is empty, all detected CI systems are merged.
type EnvBlock struct {
	FromCI []string          `yaml:"from_ci,omitempty"`
	Vars   map[string]string `yaml:"vars,omitempty"`
}

// Service describes an external dependency that should be available during the
// pipeline (e.g. Redis, PostgreSQL). It can be ephemeral or persistent.
type Service struct {
	Name       string            `yaml:"name"`
	Plugin     string            `yaml:"plugin,omitempty"`
	Image      string            `yaml:"image,omitempty"`
	Engine     string            `yaml:"engine,omitempty"`
	Acl        []string          `yaml:"acl,omitempty"`
	Ports      []string          `yaml:"ports,omitempty"`
	Vars       map[string]string `yaml:"vars,omitempty"`
	Persistent bool              `yaml:"persistent,omitempty"`
}

// Pipeline is the top-level parsed structure.
type Pipeline struct {
	Jobs     map[string]Job     `yaml:"jobs"`
	Services map[string]Service `yaml:"services,omitempty"`
	Env      EnvBlock           `yaml:"env,omitempty"`
}

// Rule controls when a job should run based on environment conditions.
// If If evaluates to a non-empty string, the rule matches.
// When controls scheduling: "on_success" (default), "always", or "never".
type Rule struct {
	If   string `yaml:"if,omitempty"`
	When string `yaml:"when,omitempty"`
}

// Job represents a single CI job.
type Job struct {
	Needs         []string          `yaml:"needs,omitempty"`
	Plugin        string            `yaml:"plugin,omitempty"`
	Invoke        string            `yaml:"invoke,omitempty"`
	Vars          map[string]string `yaml:"vars,omitempty"`
	MemoryLimitMb int               `yaml:"memory_limit_mb,omitempty"`
	TimeoutMs     int               `yaml:"timeout_ms,omitempty"`
	Rules         []Rule            `yaml:"rules,omitempty"`
	Tags          []string          `yaml:"tags,omitempty"`
}

// Load reads file at path and unmarshals the strict pipeline YAML.
func Load(path string) (*Pipeline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return Parse(data)
}

// EvaluateRules checks if a job should run given the current environment.
// Returns true if no rules are defined (default: always run).
// If any rule has a matching If expression, the job runs.
// When is validated but currently only used for documentation.
func EvaluateRules(job Job, env map[string]string) bool {
	if len(job.Rules) == 0 {
		return true
	}

	for _, rule := range job.Rules {
		if rule.If == "" {
			// Unconditional rule; allow unless When==never
			if rule.When == "never" {
				return false
			}
			return true
		}
		// Evaluate If by checking the environment.
		// The simplest model: a non-empty value in env means the expression is true.
		if env[rule.If] != "" {
			if rule.When == "never" {
				return false
			}
			return true
		}
	}

	return false
}

// Parse unmarshals raw YAML bytes into a Pipeline and validates structural
// constraints.
func Parse(data []byte) (*Pipeline, error) {
	var p Pipeline
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	if err := validate(&p); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}
	return &p, nil
}

func validate(p *Pipeline) error {
	seen := make(map[string]struct{}, len(p.Jobs))
	for name, job := range p.Jobs {
		if _, ok := seen[name]; ok {
			return fmt.Errorf("duplicate job name %q", name)
		}
		seen[name] = struct{}{}
		if job.Plugin != "" && job.Invoke != "" {
			return fmt.Errorf("job %q: plugin and invoke are mutually exclusive", name)
		}
		if job.Plugin == "" && job.Invoke == "" {
			return fmt.Errorf("job %q: one of plugin or invoke is required", name)
		}
	}
	return nil
}
