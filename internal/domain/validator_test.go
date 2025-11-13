package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateType(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name  string
		ttype string
		want  bool
	}{
		{"valid aftn", "aftn", true},
		{"valid sita", "sita", true},
		{"valid acars", "acars", true},
		{"valid cpdlc", "cpdlc", true},
		{"invalid type", "invalid", false},
		{"empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, v.ValidateType(tt.ttype))
		})
	}
}

func TestValidator_ValidatePriority(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		priority int
		want     bool
	}{
		{"priority 1", 1, true},
		{"priority 2", 2, true},
		{"priority 3", 3, true},
		{"priority 0", 0, false},
		{"priority 4", 4, false},
		{"negative priority", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, v.ValidatePriority(tt.priority))
		})
	}
}

func TestValidator_ValidateICAOCode(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name string
		code string
		want bool
	}{
		{"valid ICAO", "ZBAA", true},
		{"valid ICAO lowercase", "zbaa", false},
		{"too short", "ZBA", false},
		{"too long", "ZBAAA", false},
		{"contains numbers", "ZB11", false},
		{"contains special chars", "ZB-A", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, v.ValidateICAOCode(tt.code))
		})
	}
}

