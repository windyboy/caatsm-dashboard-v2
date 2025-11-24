package testing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestSuite provides a base test suite with common setup/teardown
type TestSuite struct {
	suite.Suite
	Env *TestEnv
	ctx context.Context
}

// SetupSuite runs once before all tests in the suite
func (s *TestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Only setup test environment for integration tests
	// Unit tests should skip this
	if !testing.Short() {
		env, err := SetupTestEnv(s.ctx)
		if err != nil {
			s.T().Fatalf("Failed to setup test environment: %v", err)
		}
		s.Env = env
	}
}

// TearDownSuite runs once after all tests in the suite
func (s *TestSuite) TearDownSuite() {
	if s.Env != nil {
		if err := s.Env.Cleanup(s.ctx); err != nil {
			s.T().Logf("Failed to cleanup test environment: %v", err)
		}
	}
}

// SetupTest runs before each test
func (s *TestSuite) SetupTest() {
	// Clean database before each test if environment is available
	if s.Env != nil && s.Env.Pool != nil {
		if err := CleanDatabase(s.ctx, s.Env.Pool); err != nil {
			s.T().Logf("Failed to clean database: %v", err)
		}
	}
}

// Context returns the test context
func (s *TestSuite) Context() context.Context {
	if s.ctx == nil {
		s.ctx = context.Background()
	}
	return s.ctx
}
