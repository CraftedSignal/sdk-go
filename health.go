package craftedsignal

// HealthService provides health and coverage metrics.
type HealthService interface{}

type healthService struct{ t *transport }
