package requests

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
)

// tagRule reads one numeric rule (max=, min=) straight off a binding tag, so
// the test fails when the tag drifts even if the constant does not.
func tagRule(t *testing.T, typ reflect.Type, field, rule string) int {
	t.Helper()
	sf, ok := typ.FieldByName(field)
	require.True(t, ok, "field %s.%s missing", typ.Name(), field)
	for part := range strings.SplitSeq(sf.Tag.Get("binding"), ",") {
		if value, found := strings.CutPrefix(part, rule+"="); found {
			n, err := strconv.Atoi(value)
			require.NoError(t, err, "unparseable %s tag on %s", rule, field)
			return n
		}
	}
	t.Fatalf("rule %s missing on %s", rule, field)
	return 0
}

func TestVideoBindingLimitsMatchConstants(t *testing.T) {
	tests := []struct {
		name  string
		typ   reflect.Type
		field string
		rule  string
		want  int
	}{
		{
			name:  "preview max equals VideoPreviewMax",
			typ:   reflect.TypeOf(PreviewVideoCategories{}),
			field: "PreviewK",
			rule:  "max",
			want:  common.VideoPreviewMax,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tagRule(t, tt.typ, tt.field, tt.rule))
		})
	}
}
