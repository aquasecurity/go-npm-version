package npm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConstraints(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{">= 1.1", false},
		{">40.50.60, < 50.70", false},
		{"2.0", false},
		{"2.3.5-20161202202307-sha.e8fc5e5", false},
		{">= bar", true},
		{"BAR >= 1.2.3", true},

		// Test with commas separating AND
		{">= 1.2.3, < 2.0", false},
		{">= 1.2.3, < 2.0 || => 3.0, < 4", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := NewConstraints(tt.input)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConstraint_Check(t *testing.T) {
	for _, tc := range autoGeneratedTests {
		t.Run(fmt.Sprintf("%s vs %s", tc.constraint, tc.version), func(t *testing.T) {
			c, err := NewConstraints(tc.constraint)
			require.NoError(t, err)

			v, err := NewVersion(tc.version)
			require.NoError(t, err)

			got := c.Check(v)
			assert.Equal(t, tc.want, got)
		})
	}
}