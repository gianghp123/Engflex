package tests

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"

	"engflex-api/internal/database/models"
)

// UsageCounter's composite primary key is the identity the atomic Reserve
// upserts against: ON CONFLICT (user_id, counter_key, period_key). A
// single-column or surrogate PK would silently change the conflict target
// and split one logical counter into many rows.
func TestUsageCounterCompositePrimaryKey(t *testing.T) {
	s, err := schema.Parse(&models.UsageCounter{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	require.Len(t, s.PrimaryFields, 3,
		"the counter identity is (user, counter, period) — no surrogate id")
	names := make([]string, 0, len(s.PrimaryFields))
	for _, f := range s.PrimaryFields {
		names = append(names, f.DBName)
	}
	assert.ElementsMatch(t, []string{"user_id", "counter_key", "period_key"}, names)
}

// The (user_id, idempotency_key) unique index is the exactly-once guard: a
// retried Reserve with the same base key must hit the existing row, never
// insert a second charge. GORM's `uniqueIndex` tag creates a single-column
// index unless the SAME index name is tagged on BOTH fields — assert via
// ParseIndexes (through LookIndex), not TagSettings, so a single-column
// drift fails here.
func TestUsageEventCompositeIdempotencyIndex(t *testing.T) {
	s, err := schema.Parse(&models.UsageEvent{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	idx := s.LookIndex("idx_usage_events_user_idem")
	require.NotNil(t, idx,
		"UserID and IdempotencyKey must share one uniqueIndex name")
	assert.Equal(t, "UNIQUE", idx.Class)
	names := make([]string, 0, len(idx.Fields))
	for _, f := range idx.Fields {
		require.NotNil(t, f.Field)
		names = append(names, f.Field.DBName)
	}
	assert.ElementsMatch(t, []string{"user_id", "idempotency_key"}, names,
		"a single-column index would let the same key charge twice")
}
