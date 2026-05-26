package pipeline

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBasic(t *testing.T) {
	yaml := `
jobs:
  lint:
    plugin: docker://golangci/golangci-lint:v1.54
  build:
    needs: [lint]
    plugin: docker://golang:1.22
    vars:
      CGO_ENABLED: "0"
`
	p, err := Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Len(t, p.Jobs, 2)
	assert.Equal(t, "docker://golangci/golangci-lint:v1.54", p.Jobs["lint"].Plugin)
	assert.Len(t, p.Jobs["build"].Needs, 1)
	assert.Equal(t, "lint", p.Jobs["build"].Needs[0])
}

func TestParseRule(t *testing.T) {
	yaml := `
jobs:
  release:
    plugin: docker://alpine
    rules:
      - if: "CI_COMMIT_TAG"
      - if: "CI_COMMIT_BRANCH"
        when: "never"
`
	p, err := Parse([]byte(yaml))
	require.NoError(t, err)
	job, ok := p.Jobs["release"]
	require.True(t, ok)
	assert.Len(t, job.Rules, 2)
	assert.Equal(t, "CI_COMMIT_TAG", job.Rules[0].If)
	assert.Equal(t, "CI_COMMIT_BRANCH", job.Rules[1].If)
	assert.Equal(t, "never", job.Rules[1].When)
}

func TestEvaluateRulesNoRules(t *testing.T) {
	j := Job{Plugin: "docker://alpine"}
	assert.True(t, EvaluateRules(j, nil))
}

func TestEvaluateRulesIfMatch(t *testing.T) {
	j := Job{
		Plugin: "docker://alpine",
		Rules:  []Rule{{If: "CI_COMMIT_TAG"}},
	}
	assert.False(t, EvaluateRules(j, map[string]string{"CI_COMMIT_TAG": ""}))
	assert.True(t, EvaluateRules(j, map[string]string{"CI_COMMIT_TAG": "v1.0.0"}))
}

func TestEvaluateRulesWhenNever(t *testing.T) {
	j := Job{
		Plugin: "docker://alpine",
		Rules:  []Rule{{If: "CI_COMMIT_TAG", When: "never"}},
	}
	assert.False(t, EvaluateRules(j, map[string]string{"CI_COMMIT_TAG": "v1.0.0"}))
}

func TestEvaluateRulesEmptyIfMatches(t *testing.T) {
	j := Job{
		Plugin: "docker://alpine",
		Rules:  []Rule{{If: "", When: "on_success"}},
	}
	assert.True(t, EvaluateRules(j, nil))
}

func TestEvaluateRulesEmptyIfWhenNever(t *testing.T) {
	j := Job{
		Plugin: "docker://alpine",
		Rules:  []Rule{{If: "", When: "never"}},
	}
	assert.False(t, EvaluateRules(j, nil))
}

func TestEvaluateRulesNoMatch(t *testing.T) {
	j := Job{
		Plugin: "docker://alpine",
		Rules:  []Rule{{If: "MISSING_VAR"}},
	}
	assert.False(t, EvaluateRules(j, map[string]string{}))
}

func TestParseEmpty(t *testing.T) {
	p, err := Parse([]byte(`jobs:`))
	require.NoError(t, err)
	assert.Len(t, p.Jobs, 0)
}

func TestParseInvalidYAML(t *testing.T) {
	_, err := Parse([]byte(`jobs: [`))
	require.Error(t, err)
}

func TestParseWithInvoke(t *testing.T) {
	yaml := `
jobs:
  test:
    invoke: echo hello
`
	p, err := Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Equal(t, "echo hello", p.Jobs["test"].Invoke)
	assert.Equal(t, "", p.Jobs["test"].Plugin)
}

func TestParseWithServices(t *testing.T) {
	yaml := `
jobs:
  build:
    plugin: docker://alpine
services:
  redis:
    image: redis:7
    ports:
      - "6379:6379"
  postgres:
    image: postgres:16
    persistent: true
`
	p, err := Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Len(t, p.Jobs, 1)
	assert.Len(t, p.Services, 2)
	assert.Equal(t, "redis:7", p.Services["redis"].Image)
	assert.True(t, p.Services["postgres"].Persistent)
	assert.Contains(t, p.Services["redis"].Ports, "6379:6379")
}

func TestParseWithEnv(t *testing.T) {
	yaml := `
env:
  from_ci:
    - CI_COMMIT_SHA
    - CI_COMMIT_BRANCH
  vars:
    FOO: bar
jobs:
  build:
    plugin: docker://alpine
`
	p, err := Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Len(t, p.Env.FromCI, 2)
	assert.Equal(t, p.Env.Vars["FOO"], "bar")
}

func TestParseDuplicateJob(t *testing.T) {
	yaml := `
jobs:
  build:
    plugin: docker://alpine
  build:
    invoke: echo dup
`
	_, err := Parse([]byte(yaml))
	require.Error(t, err)
}

func TestParsePluginAndInvokeMutuallyExclusive(t *testing.T) {
	yaml := `
jobs:
  bad:
    plugin: docker://alpine
    invoke: echo hello
`
	_, err := Parse([]byte(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestParseNeitherPluginNorInvoke(t *testing.T) {
	yaml := `
jobs:
  bad:
    needs: []
`
	_, err := Parse([]byte(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "one of plugin or invoke is required")
}

func TestLoad_Nonexistent(t *testing.T) {
	_, err := Load("/nonexistent/path/.coreci.yml")
	require.Error(t, err)
}

func TestLoad_Valid(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/.coreci.yml"
	require.NoError(t, os.WriteFile(path, []byte("jobs:\n  test:\n    invoke: echo hi\n"), 0644))
	p, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "echo hi", p.Jobs["test"].Invoke)
}

func TestParseJobWithMemoryLimit(t *testing.T) {
	yaml := `
jobs:
  heavy:
    plugin: docker://alpine
    memory_limit_mb: 512
    timeout_ms: 60000
`
	p, err := Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Equal(t, 512, p.Jobs["heavy"].MemoryLimitMb)
	assert.Equal(t, 60000, p.Jobs["heavy"].TimeoutMs)
}
