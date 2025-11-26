package health

import (
	"sync"
	"testing"

	"go.uber.org/zap"
)

// TestPolicyManagerConcurrency tests that PolicyManager is safe for concurrent use.
func TestPolicyManagerConcurrency(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(logger)

	// Run concurrent reads and writes
	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Concurrent readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = pm.GetPolicy("meilisearch")
				_ = pm.GetPoliciesForComponents([]string{"meilisearch", "postgresql"})
				_ = pm.ShouldShowBanner([]string{"redis"})
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				policy := DegradationPolicy{
					Component:     "test_component",
					Action:        ActionContinue,
					Fallback:      "test fallback",
					ShowBanner:    true,
					BannerMessage: "test message",
				}
				pm.SetPolicy(policy)
			}
		}(i)
	}

	// Mixed operations
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if j%2 == 0 {
					pm.SetPolicy(DegradationPolicy{
						Component:     "mixed_test",
						Action:        ActionFallbackToDB,
						Fallback:      "mixed fallback",
						ShowBanner:    false,
						BannerMessage: "",
					})
				} else {
					_, _ = pm.GetPolicy("mixed_test")
					_ = pm.GetBannerMessages([]string{"meilisearch", "mixed_test"})
				}
			}
		}()
	}

	wg.Wait()

	// Verify we can still read after all concurrent operations
	policy, ok := pm.GetPolicy("meilisearch")
	if !ok {
		t.Error("Expected to find meilisearch policy")
	}
	if policy.Component != "meilisearch" {
		t.Errorf("Expected component 'meilisearch', got '%s'", policy.Component)
	}
}

// TestPolicyManagerGetPolicy tests the GetPolicy method.
func TestPolicyManagerGetPolicy(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(logger)

	tests := []struct {
		name      string
		component string
		wantOK    bool
	}{
		{
			name:      "existing policy - meilisearch",
			component: "meilisearch",
			wantOK:    true,
		},
		{
			name:      "existing policy - postgresql",
			component: "postgresql",
			wantOK:    true,
		},
		{
			name:      "existing policy - redis",
			component: "redis",
			wantOK:    true,
		},
		{
			name:      "non-existing policy",
			component: "nonexistent",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, ok := pm.GetPolicy(tt.component)
			if ok != tt.wantOK {
				t.Errorf("GetPolicy() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && policy.Component != tt.component {
				t.Errorf("GetPolicy() component = %v, want %v", policy.Component, tt.component)
			}
		})
	}
}

// TestPolicyManagerSetPolicy tests the SetPolicy method.
func TestPolicyManagerSetPolicy(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(logger)

	customPolicy := DegradationPolicy{
		Component:     "custom_component",
		Action:        ActionDisable,
		Fallback:      "custom fallback",
		ShowBanner:    true,
		BannerMessage: "Custom message",
	}

	pm.SetPolicy(customPolicy)

	policy, ok := pm.GetPolicy("custom_component")
	if !ok {
		t.Fatal("Expected to find custom_component policy")
	}

	if policy.Component != customPolicy.Component {
		t.Errorf("Component = %v, want %v", policy.Component, customPolicy.Component)
	}
	if policy.Action != customPolicy.Action {
		t.Errorf("Action = %v, want %v", policy.Action, customPolicy.Action)
	}
	if policy.Fallback != customPolicy.Fallback {
		t.Errorf("Fallback = %v, want %v", policy.Fallback, customPolicy.Fallback)
	}
}

// TestPolicyManagerGetPoliciesForComponents tests getting multiple policies.
func TestPolicyManagerGetPoliciesForComponents(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(logger)

	components := []string{"meilisearch", "postgresql", "nonexistent"}
	policies := pm.GetPoliciesForComponents(components)

	if len(policies) != 2 {
		t.Errorf("Expected 2 policies, got %d", len(policies))
	}

	if _, ok := policies["meilisearch"]; !ok {
		t.Error("Expected meilisearch policy")
	}
	if _, ok := policies["postgresql"]; !ok {
		t.Error("Expected postgresql policy")
	}
	if _, ok := policies["nonexistent"]; ok {
		t.Error("Did not expect nonexistent policy")
	}
}

// TestValidatePolicy tests policy validation.
func TestValidatePolicy(t *testing.T) {
	tests := []struct {
		name    string
		policy  DegradationPolicy
		wantErr bool
	}{
		{
			name: "valid policy",
			policy: DegradationPolicy{
				Component: "test",
				Action:    ActionFallbackToDB,
			},
			wantErr: false,
		},
		{
			name: "empty component",
			policy: DegradationPolicy{
				Component: "",
				Action:    ActionFallbackToDB,
			},
			wantErr: true,
		},
		{
			name: "invalid action",
			policy: DegradationPolicy{
				Component: "test",
				Action:    "invalid_action",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePolicy(tt.policy)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

