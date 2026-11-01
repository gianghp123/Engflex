package tests

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/exams/dtos/responses"
	"engflex-api/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

func TestPostgresExamCatalogCounts(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated PostgreSQL test database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlog.Discard})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })
	ctx := context.Background()
	prefix := uuid.NewString()
	tests := []struct {
		name  string
		count int
	}{{"empty", 0}, {"one", 1}, {"beyond page size", 13}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat := &models.ExamCategory{Slug: uuid.NewString(), Title: prefix + tt.name, Type: enums.ExamTypeTOEIC}
			require.NoError(t, tx.Create(cat).Error)
			for i := 0; i < tt.count; i++ {
				require.NoError(t, tx.Create(&models.ExamTest{CategoryID: cat.ID, SourceKey: uuid.NewString(), Slug: uuid.NewString(), Title: "Count probe", Content: datatypes.JSON(`{}`)}).Error)
			}
			for _, filter := range []enums.ExamType{"", enums.ExamTypeTOEIC} {
				rows, err := repositories.NewExamCategoryRepository(tx).List(ctx, filter)
				require.NoError(t, err)
				var found *models.ExamCategoryCatalog
				for i, row := range rows {
					if i > 0 {
						require.LessOrEqual(t, rows[i-1].Title, row.Title)
						if rows[i-1].Title == row.Title {
							require.Less(t, rows[i-1].ID, row.ID)
						}
					}
					if row.ID == cat.ID {
						found = row
					}
				}
				require.NotNil(t, found)
				require.EqualValues(t, tt.count, found.TestsCount)
				var dto responses.CategoryCatalog
				require.NoError(t, utils.Map(&dto, found))
				encoded, err := json.Marshal(dto)
				require.NoError(t, err)
				var wire map[string]any
				require.NoError(t, json.Unmarshal(encoded, &wire))
				require.Equal(t, cat.ID, wire["id"])
				require.Equal(t, cat.Title, wire["title"])
				require.Equal(t, float64(tt.count), wire["testsCount"])
				require.NotContains(t, wire, "category")
			}
			// Shared joined categories must never select a nonexistent count column.
			if tt.count > 0 {
				listed, err := repositories.NewExamTestRepository(tx).ListByCategory(ctx, cat.ID, 1, 0)
				require.NoError(t, err)
				require.Len(t, listed, 1)
				detail, err := repositories.NewExamTestRepository(tx).GetByID(ctx, listed[0].ID)
				require.NoError(t, err)
				require.Equal(t, cat.ID, detail.Category.ID)
			}
		})
	}
}
