package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"engflex-api/internal/utils"
)

// TestTidyTermAndTermKey pins the identity rules on the Go side. The same table
// is asserted against the database's generated term_key in Task 1 Step 9, so the
// two derivations cannot drift apart.
func TestTidyTermAndTermKey(t *testing.T) {
	tests := []struct {
		name string
		in   string
		tidy string
		key  string
	}{
		{"trims", "  bank  ", "bank", "bank"},
		{"preserves display case, folds the key", "Bank", "Bank", "bank"},
		{"collapses inner whitespace", "give   up", "give up", "give up"},
		{"collapses tabs and newlines", "give\t\nup", "give up", "give up"},
		{"camel case term", "iPhone", "iPhone", "iphone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.tidy, utils.TidyTerm(tt.in))
			assert.Equal(t, tt.key, utils.TermKey(tt.in))
		})
	}
}
