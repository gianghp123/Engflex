package tests

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"

	"engflex-api/internal/database/models"
)

func TestVocabularyRelationsAreReadOnly(t *testing.T) {
	s, err := schema.Parse(&models.VocabularyDeck{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	field, ok := s.FieldsByName["Category"]
	require.True(t, ok, "Category must remain a parsed field")
	assert.True(t, field.Readable)
	assert.False(t, field.Creatable)
	assert.False(t, field.Updatable)
}

func TestDeckItemDeckRelationIsReadOnly(t *testing.T) {
	s, err := schema.Parse(&models.VocabularyDeckItem{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	field, ok := s.FieldsByName["Deck"]
	require.True(t, ok, "Deck must remain a parsed field")
	assert.True(t, field.Readable)
	assert.False(t, field.Creatable)
	assert.False(t, field.Updatable)
	rel, ok := s.Relationships.Relations["Deck"]
	require.True(t, ok)
	require.Len(t, rel.References, 1)
	assert.Equal(t, "deck_id", rel.References[0].ForeignKey.DBName)
}

func TestItemUserStateRelationIsReadOnly(t *testing.T) {
	s, err := schema.Parse(&models.VocabularyItem{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	field, ok := s.FieldsByName["UserState"]
	require.True(t, ok, "UserState must remain a parsed field")
	assert.True(t, field.Readable)
	assert.False(t, field.Creatable)
	assert.False(t, field.Updatable)
	rel, ok := s.Relationships.Relations["UserState"]
	require.True(t, ok)
	require.Len(t, rel.References, 1)
	assert.Equal(t, "item_id", rel.References[0].ForeignKey.DBName)
}

func TestDeckItemItemRelationIsReadOnly(t *testing.T) {
	s, err := schema.Parse(&models.VocabularyDeckItem{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	field, ok := s.FieldsByName["Item"]
	require.True(t, ok, "Item must remain a parsed field")
	assert.True(t, field.Readable)
	assert.False(t, field.Creatable)
	assert.False(t, field.Updatable)
	rel, ok := s.Relationships.Relations["Item"]
	require.True(t, ok)
	require.Len(t, rel.References, 1)
	assert.Equal(t, "item_id", rel.References[0].ForeignKey.DBName)
}

func TestDeckItemsRelationIsReadOnly(t *testing.T) {
	s, err := schema.Parse(&models.VocabularyDeck{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	field, ok := s.FieldsByName["Items"]
	require.True(t, ok, "Items must remain a parsed field")
	assert.True(t, field.Readable)
	assert.False(t, field.Creatable)
	assert.False(t, field.Updatable)
	rel, ok := s.Relationships.Relations["Items"]
	require.True(t, ok)
	require.Len(t, rel.References, 1)
	assert.Equal(t, "deck_id", rel.References[0].ForeignKey.DBName)
}
