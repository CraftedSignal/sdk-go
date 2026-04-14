package craftedsignal

import "time"

// Me contains authentication information for the current token.
type Me struct {
	Company    string   `json:"company"`
	APIKeyName string   `json:"api_key_name"`
	Scopes     []string `json:"scopes"`
}

// Detection represents a detection rule.
// yaml tags allow the CLI to unmarshal YAML files directly into this type.
type Detection struct {
	ID          string   `yaml:"id,omitempty"           json:"id,omitempty"`
	Title       string   `yaml:"title"                  json:"title"`
	Description string   `yaml:"description,omitempty"  json:"description,omitempty"`
	Platform    string   `yaml:"platform"               json:"platform"`
	Query       string   `yaml:"query,omitempty"        json:"query,omitempty"`
	SigmaSource string   `yaml:"sigma_source,omitempty" json:"sigma_source,omitempty"`
	Severity    string   `yaml:"severity,omitempty"     json:"severity,omitempty"`
	Kind        string   `yaml:"kind,omitempty"         json:"kind,omitempty"`
	Enabled     bool     `yaml:"enabled"                json:"enabled"`
	Frequency   string   `yaml:"frequency,omitempty"    json:"frequency,omitempty"`
	Period      string   `yaml:"period,omitempty"       json:"period,omitempty"`
	Tactics     []string `yaml:"tactics,omitempty"      json:"tactics,omitempty"`
	Techniques  []string `yaml:"techniques,omitempty"   json:"techniques,omitempty"`
	Tags        []string `yaml:"tags,omitempty"         json:"tags,omitempty"`
	Groups      []string `yaml:"groups,omitempty"       json:"groups,omitempty"`
	Tests       *DetectionTests `yaml:"tests,omitempty" json:"tests,omitempty"`

	// Read-only fields set by the API (ignored when parsing YAML files)
	TestStatus string     `yaml:"-" json:"test_status,omitempty"`
	Version    int        `yaml:"-" json:"version,omitempty"`
	UpdatedAt  *time.Time `yaml:"-" json:"updated_at,omitempty"`

	// AI generation metadata (read-only)
	AIGenerated    bool     `yaml:"-" json:"ai_generated,omitempty"`
	AIQualityScore *float64 `yaml:"-" json:"ai_quality_score,omitempty"`
}

// DetectionTests contains test cases for a detection rule.
type DetectionTests struct {
	Positive []DetectionTest     `yaml:"positive,omitempty" json:"positive,omitempty"`
	Negative []DetectionTest     `yaml:"negative,omitempty" json:"negative,omitempty"`
	Simulate []SimulationBinding `yaml:"simulate,omitempty" json:"simulate,omitempty"`
}

// DetectionTest is a single positive or negative test case.
type DetectionTest struct {
	Name        string                   `yaml:"name"                  json:"name"`
	Description string                   `yaml:"description,omitempty" json:"description,omitempty"`
	Data        []map[string]interface{} `yaml:"data,omitempty"        json:"data,omitempty"`
	JSON        string                   `yaml:"json,omitempty"        json:"json,omitempty"`
}

// SimulationBinding declares an explicit simulation binding for a detection rule.
type SimulationBinding struct {
	Technique string   `yaml:"technique"          json:"technique"`
	Adapters  []string `yaml:"adapters,omitempty" json:"adapters,omitempty"`
	Expected  bool     `yaml:"expected"           json:"expected"`
}

// SyncStatus contains version tracking info for all rules in the workspace.
type SyncStatus struct {
	Rules []SyncStatusRule `json:"rules"`
}

// SyncStatusRule is the sync state for a single rule.
type SyncStatusRule struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Groups    []string  `json:"groups"`
	Hash      string    `json:"hash"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ImportRequest is the payload for POST /api/v1/detections/import.
type ImportRequest struct {
	Rules     []Detection `json:"rules"`
	Message   string      `json:"message"`
	Mode      string      `json:"mode"`
	Atomic    *bool       `json:"atomic,omitempty"`
	SkipTests bool        `json:"skip_tests,omitempty"`
}

// ImportResponse is the result of an import operation.
type ImportResponse struct {
	Success    bool           `json:"success"`
	RolledBack bool           `json:"rolled_back,omitempty"`
	Results    []ImportResult `json:"results"`
	Created    int            `json:"created"`
	Updated    int            `json:"updated"`
	Unchanged  int            `json:"unchanged"`
	Conflicts  int            `json:"conflicts"`
	Errors     int            `json:"errors"`
	StatusCode int            `json:"-"`
}

// ImportResult is the per-rule outcome of an import.
type ImportResult struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Action  string `json:"action"`
	Error   string `json:"error,omitempty"`
	Version int    `json:"version"`
}

// DiffResult is the result of comparing a local rule against the remote version.
type DiffResult struct {
	HasDiff bool   `json:"has_diff"`
	Diff    string `json:"diff"`
}

// DeployRequest is the payload for POST /api/v1/detections/deploy.
type DeployRequest struct {
	DetectionIDs  []string `json:"detection_ids"`
	OverrideTests bool     `json:"override_tests"`
}

// DeployResponse is the result of a deploy operation.
type DeployResponse struct {
	Results  []DeployResult `json:"results"`
	Deployed int            `json:"deployed"`
	Failed   int            `json:"failed"`
}

// DeployResult is the per-rule outcome of a deploy.
type DeployResult struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Action string `json:"action"`
	Error  string `json:"error,omitempty"`
}

// TestJob holds the result of starting async test execution.
type TestJob struct {
	Results []TestStartResult `json:"results"`
	Started int               `json:"started"`
	Skipped int               `json:"skipped"`
	Errors  int               `json:"errors"`
}

// TestStartResult is the per-rule outcome of triggering a test.
type TestStartResult struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Action     string `json:"action"`
	WorkflowID string `json:"workflow_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

// TestResponse is the polled status of test execution.
type TestResponse struct {
	Results []TestResult `json:"results"`
	Passed  int          `json:"passed"`
	Failed  int          `json:"failed"`
	Pending int          `json:"pending"`
}

// TestResult is the per-rule test status.
type TestResult struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	TestStatus  string        `json:"test_status"`
	FailedTests []TestFailure `json:"failed_tests,omitempty"`
}

// TestFailure describes a single failed test case.
type TestFailure struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Error   string `json:"error,omitempty"`
	Matches int    `json:"matches,omitempty"`
}

// GenerateRequest is the payload for POST /api/v1/detections/generate.
type GenerateRequest struct {
	Description string `json:"description"`
	Platform    string `json:"platform"`
	SigmaYAML   string `json:"sigma_yaml,omitempty"`
}

// GenerateJob holds the workflow ID for an in-progress generation.
type GenerateJob struct {
	WorkflowID string `json:"workflow_id"`
	Status     string `json:"status"`
}

// GenerateResult is the terminal result of an AI generation workflow.
type GenerateResult struct {
	Status   string      `json:"status"`
	Progress string      `json:"progress,omitempty"`
	Error    string      `json:"error,omitempty"`
	Rules    []Detection `json:"rules,omitempty"`
}

// Approval represents a pending deployment approval.
type Approval struct {
	ID            string        `json:"id"`
	CompanyID     uint64        `json:"company_id"`
	DeploymentID  string        `json:"deployment_id"`
	RuleID        string        `json:"rule_id"`
	Status        string        `json:"status"`
	ImpactSummary ImpactSummary `json:"impact_summary"`
	CreatedAt     time.Time     `json:"created_at"`
}

// ImpactSummary describes the projected impact of a deployment.
type ImpactSummary struct {
	TargetPlatforms []string `json:"target_platforms"`
	AffectedIndexes []string `json:"affected_indexes"`
	ProjectedAlerts int      `json:"projected_alerts"`
	RiskLevel       string   `json:"risk_level"`
	QueryLatencyMs  int      `json:"query_latency_ms"`
}

// SimulationRun represents a simulation execution.
type SimulationRun struct {
	ID            string             `json:"id"`
	TechniqueID   string             `json:"technique_id"`
	TechniqueName string             `json:"technique_name"`
	Adapter       string             `json:"adapter"`
	ExecMode      string             `json:"exec_mode"`
	Target        string             `json:"target"`
	OS            string             `json:"os"`
	Status        string             `json:"status"`
	Results       []SimulationResult `json:"results,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
}

// SimulationResult is a single detection match within a simulation run.
type SimulationResult struct {
	MatchMethod     string  `json:"match_method"`
	Matched         bool    `json:"matched"`
	MatchConfidence float64 `json:"match_confidence"`
	MatchCount      int     `json:"match_count"`
}

// CreateSimulationRequest is the payload for POST /api/v1/simulations/runs.
type CreateSimulationRequest struct {
	TechniqueID string `json:"technique_id"`
	Adapter     string `json:"adapter"`
	Target      string `json:"target"`
	ExecMode    string `json:"exec_mode,omitempty"`
}

// VerifyJob holds state for an in-progress MITRE correlation.
type VerifyJob struct {
	RunID  string `json:"run_id"`
	Status string `json:"status"`
}

// VerifyResult is the terminal result of a simulation verification.
type VerifyResult struct {
	Status  string             `json:"status"`
	Results []SimulationResult `json:"results,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// CoverageReport summarises detection coverage across MITRE techniques.
type CoverageReport struct {
	Total    int     `json:"total"`
	Covered  int     `json:"covered"`
	Coverage float64 `json:"coverage"`
}

// CoverageGap is a MITRE technique with no covering detection.
type CoverageGap struct {
	TechniqueID   string `json:"technique_id"`
	TechniqueName string `json:"technique_name"`
	Tactic        string `json:"tactic"`
}

// HealthMetrics contains company-wide detection health data.
type HealthMetrics struct {
	TotalRules   int     `json:"total_rules"`
	PassingRules int     `json:"passing_rules"`
	FailingRules int     `json:"failing_rules"`
	NoTestRules  int     `json:"no_test_rules"`
	HealthScore  float64 `json:"health_score"`
}

// NoiseBudget contains alert fatigue budget metrics.
type NoiseBudget struct {
	DailyBudget   int     `json:"daily_budget"`
	CurrentAlerts int     `json:"current_alerts"`
	Utilisation   float64 `json:"utilisation"`
}

// APIKey is a managed API key resource (from GET /api/v1/api-keys).
// Not to be confused with Token, which is the bearer secret passed to NewClient.
type APIKey struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	KeyPrefix       string     `json:"key_prefix"`
	Scopes          []string   `json:"scopes"`
	RateLimit       int        `json:"rate_limit"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	AllowedTargets  []uint64   `json:"allowed_targets,omitempty"`
	AllowedGroups   []uint64   `json:"allowed_groups,omitempty"`
	AllowedPrefixes []string   `json:"allowed_prefixes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateAPIKeyRequest is the payload for POST /api/v1/api-keys.
type CreateAPIKeyRequest struct {
	Name            string     `json:"name"`
	Scopes          []string   `json:"scopes"`
	RateLimit       int        `json:"rate_limit,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	AllowedTargets  []uint64   `json:"allowed_targets,omitempty"`
	AllowedGroups   []uint64   `json:"allowed_groups,omitempty"`
	AllowedPrefixes []string   `json:"allowed_prefixes,omitempty"`
}

// APIKeyWithSecret includes the plaintext key, returned only on creation.
type APIKeyWithSecret struct {
	APIKey
	PlaintextKey string `json:"key"`
}

// ProgressFunc is called on each poll tick for async high-level helpers.
// status is the current workflow status string; pct is 0-100 if known, -1 if unknown.
// A nil ProgressFunc is safe — polling proceeds silently.
type ProgressFunc func(status string, pct int)
