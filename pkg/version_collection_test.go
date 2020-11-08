package npm

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollection(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		want     []string
	}{
		{
			name: "happy path",
			versions: []string{
				"1.1.1",
				"1.0.0",
				"1.2.0",
				"1.0.0-beta",
				"2.0.0",
				"0.7.1",
				"1.0.0-alpha",
			},
			want: []string{
				"0.7.1",
				"1.0.0-alpha",
				"1.0.0-beta",
				"1.0.0",
				"1.1.1",
				"1.2.0",
				"2.0.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := make([]Version, len(tt.versions))
			for i, raw := range tt.versions {
				v, err := NewVersion(raw)
				require.NoError(t, err)
				versions[i] = v
			}

			sort.Sort(Collection(versions))

			got := make([]string, len(versions))
			for i, v := range versions {
				got[i] = v.String()
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
