package tests

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"

	"engflex-api/internal/database/models"
)

// The Scenario relation on Conversation is read-only by construction.
//
// Without a write barrier GORM treats a populated Scenario as an association
// to save: writing a conversation that came back from ListByUser with its
// scenario attached silently UPDATEs or INSERTs into the scenarios catalog —
// err == nil, RowsAffected unchanged, and nothing in the logs says so. The
// failure mode is the same class as the turn-feedback duplicate the
// ConversationTurn test pins: a shared catalog row mutated by a per-user
// write that only meant to touch conversations.
//
// "->" is the only fix that keeps both halves: the field stays a registered
// relationship (so Joins("Scenario") populates it and the Scenario__* aliases
// still scan, with nil for free talk and deleted scenarios) while
// SelectAndOmitColumns marks it excluded from create and update. This test
// asserts both halves from the parsed schema, so a "simplification" back to
// a bare foreignKey tag fails here. Do not simplify to a bare
// `gorm:"foreignKey:ScenarioID;references:ID"` tag: that restores write
// permission and lets conversation writes auto-save the scenario.
func TestConversationScenarioRelationIsReadOnly(t *testing.T) {
	s, err := schema.Parse(&models.Conversation{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	field, ok := s.FieldsByName["Scenario"]
	require.True(t, ok, "Scenario must remain a parsed field")

	t.Run("readable so the join can populate it", func(t *testing.T) {
		assert.True(t, field.Readable,
			"Readable must stay true or Joins(\"Scenario\") stops scanning into it")
	})

	t.Run("not creatable or updatable so the relation is never saved", func(t *testing.T) {
		assert.False(t, field.Creatable,
			"a creatable Scenario makes GORM auto-save the scenario catalog row on insert")
		assert.False(t, field.Updatable,
			"an updatable Scenario makes GORM auto-save on update too")
	})

	t.Run("keeps enough permission to register as a relationship", func(t *testing.T) {
		// This is the assertion that separates "->" from "-". The registration
		// gate in schema.go is:
		//     DataType == "" && GORMDataType == "" && (Creatable || Updatable || Readable)
		// so Readable alone is what keeps the field in relationshipFields. "-"
		// clears all three permissions, dropping the relation:
		// Joins("Scenario") then resolves to nothing, the LEFT JOIN vanishes
		// from the SQL, and every conversation comes back with a nil
		// Scenario — roleplay rows silently lose their title and persona.
		// Verified against the parsed schema, not inferred.
		assert.True(t, field.Creatable || field.Updatable || field.Readable,
			"all permissions false unregisters the relation; \"->\" must keep Readable true")
	})

	t.Run("still a registered relationship", func(t *testing.T) {
		rel, ok := s.Relationships.Relations["Scenario"]
		require.True(t, ok,
			"Joins(\"Scenario\") resolves through Relationships.Relations; without it the join silently degrades")
		// The belongs-to pair resolves to conversations.scenario_id =
		// scenarios.id, which is the ON clause the history list needs. The
		// foreign key lives on the conversation, so OwnPrimaryKey is false.
		require.Len(t, rel.References, 1)
		assert.False(t, rel.References[0].OwnPrimaryKey)
		assert.Equal(t, "id", rel.References[0].PrimaryKey.DBName)
		assert.Equal(t, "scenario_id", rel.References[0].ForeignKey.DBName)
	})
}
