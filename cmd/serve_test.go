package cmd

import (
	"testing"
)

func TestServeCompatFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{"default", []string{"serve"}, false},
		{"compat enabled", []string{"serve", "--compat"}, true},
		{"compat disabled", []string{"serve", "--compat=false"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewServeCmd()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			compat, err := cmd.Flags().GetBool("compat")
			if err != nil {
				t.Fatalf("Failed to get compat flag: %v", err)
			}

			if compat != tt.expected {
				t.Errorf("Expected compat=%v, got %v", tt.expected, compat)
			}
		})
	}
}
