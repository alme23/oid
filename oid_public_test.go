// oid/public_test.go
package oid_test

import (
	"testing"

	"github.com/alme23/oid"
)

// ============================================
// ТЕСТЫ
// ============================================

func TestParseOID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid", "1.3.6.1", false},
		{"Invalid", "invalid", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := oid.ParseOID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOID(%q) error = %v, wantErr %v",
					tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestOIDString(t *testing.T) {
	oid := oid.MustParseOID("1.3.6.1.4.1")
	if oid.String() != "1.3.6.1.4.1" {
		t.Errorf("String() = %q, want %q", oid.String(), "1.3.6.1.4.1")
	}
}
