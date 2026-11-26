package requests

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
)

// The admin PUT binding ceiling must equal internal/common's
// MaxPlanLimitValue: drift here would let an admin write a limit the
// catalog constant forbids, or reject a legitimate one. The tag is read
// directly (not round-tripped through a value), so a silent tag change
// fails here.
func tagMax(t *testing.T, typ reflect.Type, field, rule string) int {
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

func TestPlanLimitBindingMatchesConstants(t *testing.T) {
	typ := reflect.TypeOf(PlanLimitInput{})
	require.Equal(t, common.MaxPlanLimitValue, tagMax(t, typ, "LimitValue", "max"),
		"PlanLimitInput.LimitValue max must equal MaxPlanLimitValue")
}
