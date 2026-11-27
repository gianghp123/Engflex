//go:build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
)

func newIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set; migrate the dev database first")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	require.NoError(t, err)
	return db
}

func TestPricingCatalogSchemaRoundTrip(t *testing.T) {
	db := newIntegrationDB(t)
	ctx := context.Background()
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { _ = tx.Rollback().Error })

	costs := repositories.NewActionCreditCostRepository(tx)
	rows, err := costs.List(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, 1, rows[0].Credits)
	assert.Equal(t, "system", rows[0].UpdatedBy)

	require.NoError(t, costs.UpsertBatch(ctx, []*models.ActionCreditCost{
		{Action: enums.AccessActionVoiceStart, Credits: 3, UpdatedBy: "admin_1"},
	}))
	rows, err = costs.List(ctx)
	require.NoError(t, err)
	var voice *models.ActionCreditCost
	for _, row := range rows {
		if row.Action == enums.AccessActionVoiceStart {
			voice = row
		}
	}
	require.NotNil(t, voice)
	assert.Equal(t, 3, voice.Credits)
	assert.Equal(t, "admin_1", voice.UpdatedBy)

	settings := repositories.NewTrialSettingRepository(tx)
	setting, err := settings.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, 7, setting.DurationDays)
	require.NoError(t, settings.Upsert(ctx, &models.TrialSetting{ID: true, DurationDays: 10, UpdatedBy: "admin_1"}))
	setting, err = settings.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, 10, setting.DurationDays)

	plans := repositories.NewPlanRepository(tx)
	active, err := plans.ListActive(ctx)
	require.NoError(t, err)
	require.Len(t, active, 3)
	require.NoError(t, tx.Model(&models.Plan{}).
		Where("plan = ?", enums.SubscriptionPlanTwelveMonth).
		Update("is_active", false).Error)
	active, err = plans.ListActive(ctx)
	require.NoError(t, err)
	require.Len(t, active, 2, "inactive plans never reach the public catalog")

	events := repositories.NewUsageEventRepository(tx)
	user := fmt.Sprintf("it_pricing_%d", time.Now().UnixNano())
	require.NoError(t, events.Insert(ctx, &models.UsageEvent{
		ReservationID:  "00000000-0000-0000-0000-000000000001",
		UserID:         user,
		CounterKey:     enums.CounterKindAICredit,
		PeriodKey:      "2026-11",
		Amount:         1,
		Status:         enums.UsageEventStatusReserved,
		IdempotencyKey: "it:" + user,
		RefType:        enums.UsageRefTypeConversation,
		RefID:          "00000000-0000-0000-0000-000000000002",
		Action:         enums.AccessActionVoiceStart,
		ExpiresAt:      ptrTime(time.Now().Add(time.Minute)),
	}))
	stored, err := events.ListByUser(ctx, user, 10)
	require.NoError(t, err)
	require.Len(t, stored, 1)
	assert.Equal(t, enums.AccessActionVoiceStart, stored[0].Action)
}

func ptrTime(t time.Time) *time.Time { return &t }
