package health

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// DegradationPolicy defines how the system should behave when components are degraded.
type DegradationPolicy struct {
	// Component name (e.g., "meilisearch", "postgresql", "redis")
	Component string

	// Action to take when component is down
	Action DegradationAction

	// Fallback behavior description
	Fallback string

	// Whether to show a UI banner
	ShowBanner bool

	// Banner message for users
	BannerMessage string
}

// DegradationAction defines what action to take when a component is degraded.
type DegradationAction string

const (
	// ActionFallbackToDB - Fallback to database queries (e.g., when Meilisearch is down)
	ActionFallbackToDB DegradationAction = "fallback_to_db"

	// ActionUseCache - Use cached data (e.g., when DB is slow)
	ActionUseCache DegradationAction = "use_cache"

	// ActionReadOnly - Allow read-only operations
	ActionReadOnly DegradationAction = "read_only"

	// ActionDisable - Disable the feature entirely
	ActionDisable DegradationAction = "disable"

	// ActionContinue - Continue with degraded performance
	ActionContinue DegradationAction = "continue"
)

// PolicyManager manages degradation policies and provides policy lookup.
type PolicyManager struct {
	policies map[string]DegradationPolicy
	logger   *zap.Logger
}

// NewPolicyManager creates a new policy manager with default policies.
func NewPolicyManager(logger *zap.Logger) *PolicyManager {
	pm := &PolicyManager{
		policies: make(map[string]DegradationPolicy),
		logger:   logger,
	}
	pm.setDefaultPolicies()
	return pm
}

// setDefaultPolicies sets up default degradation policies.
func (pm *PolicyManager) setDefaultPolicies() {
	pm.policies["meilisearch"] = DegradationPolicy{
		Component:     "meilisearch",
		Action:        ActionFallbackToDB,
		Fallback:      "Search will use PostgreSQL instead of Meilisearch",
		ShowBanner:    true,
		BannerMessage: "Search performance may be degraded. Using fallback search engine.",
	}

	pm.policies["postgresql"] = DegradationPolicy{
		Component:     "postgresql",
		Action:        ActionUseCache,
		Fallback:      "System will use cached data when available",
		ShowBanner:    true,
		BannerMessage: "Database connection is slow. Some features may be limited.",
	}

	pm.policies["redis"] = DegradationPolicy{
		Component:     "redis",
		Action:        ActionContinue,
		Fallback:      "Cache will be bypassed, system will continue operating",
		ShowBanner:    true,
		BannerMessage: "Caching is unavailable. Performance may be affected.",
	}
}

// GetPolicy returns the degradation policy for a component.
func (pm *PolicyManager) GetPolicy(component string) (DegradationPolicy, bool) {
	policy, ok := pm.policies[component]
	return policy, ok
}

// SetPolicy sets a custom degradation policy for a component.
func (pm *PolicyManager) SetPolicy(policy DegradationPolicy) {
	pm.policies[policy.Component] = policy
	pm.logger.Info("degradation policy set",
		zap.String("component", policy.Component),
		zap.String("action", string(policy.Action)),
	)
}

// GetPoliciesForComponents returns policies for multiple components.
func (pm *PolicyManager) GetPoliciesForComponents(components []string) map[string]DegradationPolicy {
	result := make(map[string]DegradationPolicy)
	for _, component := range components {
		if policy, ok := pm.GetPolicy(component); ok {
			result[component] = policy
		}
	}
	return result
}

// ShouldShowBanner returns true if any of the degraded components should show a banner.
func (pm *PolicyManager) ShouldShowBanner(degradedComponents []string) bool {
	for _, component := range degradedComponents {
		if policy, ok := pm.GetPolicy(component); ok && policy.ShowBanner {
			return true
		}
	}
	return false
}

// GetBannerMessages returns all banner messages for degraded components.
func (pm *PolicyManager) GetBannerMessages(degradedComponents []string) []string {
	var messages []string
	seen := make(map[string]bool)

	for _, component := range degradedComponents {
		if policy, ok := pm.GetPolicy(component); ok && policy.ShowBanner {
			if !seen[policy.BannerMessage] {
				messages = append(messages, policy.BannerMessage)
				seen[policy.BannerMessage] = true
			}
		}
	}

	return messages
}

// DegradationContext provides context about system degradation for handlers.
type DegradationContext struct {
	IsDegraded        bool
	DegradedComponents []string
	Policies          map[string]DegradationPolicy
	BannerMessages    []string
	ShouldFallback    bool // Whether to use fallback strategies
}

// GetDegradationContext creates a degradation context from health check results.
func (pm *PolicyManager) GetDegradationContext(ctx context.Context, healthService *Service) DegradationContext {
	healthResult := healthService.Check(ctx)
	degraded := healthResult.Degraded

	policies := pm.GetPoliciesForComponents(degraded)
	bannerMessages := pm.GetBannerMessages(degraded)

	shouldFallback := false
	for _, policy := range policies {
		if policy.Action == ActionFallbackToDB || policy.Action == ActionUseCache {
			shouldFallback = true
			break
		}
	}

	return DegradationContext{
		IsDegraded:        healthResult.Status == string(StatusDegraded),
		DegradedComponents: degraded,
		Policies:          policies,
		BannerMessages:    bannerMessages,
		ShouldFallback:    shouldFallback,
	}
}

// ValidatePolicy validates a degradation policy.
func ValidatePolicy(policy DegradationPolicy) error {
	if policy.Component == "" {
		return fmt.Errorf("component name is required")
	}

	validActions := map[DegradationAction]bool{
		ActionFallbackToDB: true,
		ActionUseCache:     true,
		ActionReadOnly:     true,
		ActionDisable:      true,
		ActionContinue:     true,
	}

	if !validActions[policy.Action] {
		return fmt.Errorf("invalid action: %s", policy.Action)
	}

	return nil
}

