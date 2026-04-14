# sdk-go — CraftedSignal Go SDK

[![Build](https://github.com/craftedsignal/sdk-go/actions/workflows/build.yml/badge.svg)](https://github.com/craftedsignal/sdk-go/actions/workflows/build.yml)
[![Test](https://github.com/craftedsignal/sdk-go/actions/workflows/test.yml/badge.svg)](https://github.com/craftedsignal/sdk-go/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/craftedsignal/sdk-go.svg)](https://pkg.go.dev/github.com/craftedsignal/sdk-go)

Go client for the [CraftedSignal](https://craftedsignal.io) detection-as-code API.

## Install

```bash
go get github.com/craftedsignal/sdk-go
```

Requires Go 1.22+.

## Authentication

Generate an API key in the CraftedSignal dashboard. Keys follow the `cskey_…` format.

```go
cs, err := craftedsignal.NewClient(os.Getenv("CS_TOKEN"))
if err != nil {
    log.Fatal(err)
}
me, err := cs.Me(ctx)
fmt.Println(me.Company, me.Scopes)
```

### Scopes

| Scope | Required for |
|-------|-------------|
| `rules:read` | Export, sync-status, health |
| `rules:write` | Import, diff |
| `rules:deploy` | Deploy |
| `tests:execute` | Run tests |
| `simulations:read` | Coverage, gaps, list runs |
| `simulations:write` | Create/delete runs, verify |
| `admin` | API key management |
| `detections:generate` | AI rule generation |

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    craftedsignal "github.com/craftedsignal/sdk-go"
)

func main() {
    cs, err := craftedsignal.NewClient(os.Getenv("CS_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }

    rules, err := cs.Detections.Export(context.Background(), "")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d rules\n", len(rules))
}
```

## Services

### Detections

```go
// Export rules (optionally filter by group)
rules, _ := cs.Detections.Export(ctx, "production")

// Import with atomic rollback
atomic := true
resp, _ := cs.Detections.Import(ctx, craftedsignal.ImportRequest{
    Rules: rules, Message: "sync", Mode: "upsert", Atomic: &atomic,
})

// AI generation — blocks until complete
result, _ := cs.Detections.Generate(ctx, craftedsignal.GenerateRequest{
    Description: "Detect PsExec lateral movement",
    Platform:    "splunk",
}, func(status string, _ int) { fmt.Println(status) })

// Run tests — blocks until all complete
status, _ := cs.Detections.Test(ctx, []string{"rule-id"}, nil)

// Deploy rules
deploy, _ := cs.Detections.Deploy(ctx, []string{"rule-id"}, false)
```

### Approvals

```go
approvals, _ := cs.Approvals.List(ctx)
_ = cs.Approvals.Approve(ctx, approvals[0].ID)
_ = cs.Approvals.Reject(ctx, approvals[0].ID)
```

### Simulations

```go
// Coverage report
cov, _ := cs.Simulations.Coverage(ctx)
fmt.Printf("Coverage: %.0f%%\n", cov.Coverage*100)

// Uncovered techniques
gaps, _ := cs.Simulations.Gaps(ctx)

// Run and verify simulation
run, _ := cs.Simulations.CreateRun(ctx, craftedsignal.CreateSimulationRequest{
    TechniqueID: "T1078", Adapter: "atomic", Target: "linux-host",
})
result, _ := cs.Simulations.Verify(ctx, run.ID, func(status string, _ int) {
    fmt.Println(status)
})
```

### Health

```go
m, _ := cs.Health.CompanyMetrics(ctx)
fmt.Printf("Health score: %.0f%%\n", m.HealthScore*100)

nb, _ := cs.Health.NoiseBudget(ctx)
fmt.Printf("Alert budget: %d/%d\n", nb.CurrentAlerts, nb.DailyBudget)

dead, _ := cs.Health.DeadRules(ctx)
fmt.Printf("Unused rules: %d\n", len(dead))
```

### API Keys (admin scope required)

```go
key, _ := cs.APIKeys.Create(ctx, craftedsignal.CreateAPIKeyRequest{
    Name:   "ci-pipeline",
    Scopes: []string{"rules:read", "rules:write"},
})
// Save key.PlaintextKey — it will not be shown again
fmt.Println(key.PlaintextKey)

keys, _ := cs.APIKeys.List(ctx)
_ = cs.APIKeys.Revoke(ctx, keys[0].ID)
```

## Error Handling

```go
_, err := cs.Detections.Export(ctx, "prod")

// Check for specific status codes
if errors.Is(err, craftedsignal.ErrUnauthorized) {
    // Token is invalid or expired
}
if errors.Is(err, craftedsignal.ErrNotFound) {
    // Resource does not exist
}

// Inspect the full error
var apiErr *craftedsignal.Error
if errors.As(err, &apiErr) {
    fmt.Printf("[%d %s] %s\n", apiErr.StatusCode, apiErr.Code, apiErr.Message)
}
```

## Client Options

```go
cs, err := craftedsignal.NewClient(token,
    craftedsignal.WithBaseURL("https://app.craftedsignal.io"),
    craftedsignal.WithRetry(3, craftedsignal.ExponentialBackoff),
    craftedsignal.WithVerbose(),                      // DEBUG slog output
    craftedsignal.WithLogger(slog.Default()),
    craftedsignal.WithPollInterval(2*time.Second),    // async poll cadence
    craftedsignal.WithInsecure(),                     // skip TLS verify (dev only)
    craftedsignal.WithUserAgent("my-app/1.0"),
)
```

## Async Operations

Three operations are async: **test runs**, **AI generation**, and **simulation verification**.

Each has both a low-level API (start + poll) and a high-level blocking helper:

```go
// High-level: blocks until done
result, err := cs.Detections.Generate(ctx, req, progressFn)

// Low-level: you control the loop
job, err := cs.Detections.StartGenerate(ctx, req)
for {
    poll, err := cs.Detections.PollGenerate(ctx, job.WorkflowID)
    if poll.Status == "completed" { break }
    time.Sleep(2 * time.Second)
}
```

## Mocking in Tests

Each service is an exported interface, making it straightforward to mock:

```go
type mockDetections struct{ craftedsignal.DetectionsService }

func (m *mockDetections) Export(_ context.Context, _ string) ([]craftedsignal.Detection, error) {
    return []craftedsignal.Detection{{ID: "test-rule", Title: "Test"}}, nil
}

cs, _ := craftedsignal.NewClient("token")
cs.Detections = &mockDetections{}
```

## Full API Reference

[pkg.go.dev/github.com/craftedsignal/sdk-go](https://pkg.go.dev/github.com/craftedsignal/sdk-go)

## License

MIT
