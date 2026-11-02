package services

import (
	"context"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/database/repositories/mocks"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPostgresConcurrentIPNAndRevoke(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to migrated PostgreSQL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlog.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	defer pool.Close()
	userID := "billing-" + uuid.NewString()
	ctx := context.Background()
	subs := repositories.NewSubscriptionRepository(db)
	txns := repositories.NewPaymentTransactionRepository(db)
	cfg := VNPayConfig{HashSecret: "integration-test", PayURL: "https://test.invalid/pay", ReturnURL: "https://test.invalid/return", TmnCode: "test"}
	checkout := NewVNPayService(cfg, repositories.NewTransactionManager(db), subs, txns, repositories.NewPlanRepository(db))
	session, err := checkout.BuildCheckoutURL(ctx, userID, enums.SubscriptionPlanThreeMonth, "127.0.0.1")
	require.NoError(t, err)
	var sub models.Subscription
	require.NoError(t, db.First(&sub, "user_id = ?", userID).Error)
	require.Equal(t, 3, sub.DurationMonths)
	defer func() {
		db.Where("subscription_id = ?", sub.ID).Delete(&models.PaymentTransaction{})
		db.Where("user_id = ?", userID).Delete(&models.Subscription{})
	}()
	// The plan repository must not be consulted when a pending checkout activates.
	ipn := NewVNPayService(cfg, repositories.NewTransactionManager(db), subs, txns, mocks.NewMockPlanRepository(t))
	params := map[string]string{"vnp_TxnRef": session.TxnRef, "vnp_Amount": fmt.Sprint(sub.Amount * 100), "vnp_ResponseCode": "00", "vnp_TransactionNo": "test"}
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	query.Set("vnp_SecureHash", vnpaySign(params, cfg.HashSecret))
	var wg sync.WaitGroup
	codes := make(chan string, 8)
	errs := make(chan error, 8)
	for range 8 {
		wg.Go(func() { code, e := ipn.VerifyAndHandleIPN(ctx, query); codes <- code; errs <- e })
	}
	wg.Wait()
	close(codes)
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	confirmations := 0
	for code := range codes {
		require.Contains(t, []string{vnpayRspSuccess, vnpayRspAlreadyConfirmed}, code)
		if code == vnpayRspSuccess {
			confirmations++
		}
	}
	require.Equal(t, 1, confirmations)
	active, err := subs.GetByID(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, enums.SubscriptionStatusActive, active.Status)
	require.NotNil(t, active.StartedAt)
	require.WithinDuration(t, active.StartedAt.AddDate(0, 3, 0), *active.ExpiresAt, time.Second)
	now := time.Now().UTC()
	end := now.AddDate(1, 0, 0)
	extra := &models.Subscription{UserID: userID, Plan: enums.SubscriptionPlanOneMonth, DurationMonths: 1, Status: enums.SubscriptionStatusActive, Amount: 1, Currency: "VND", StartedAt: &now, ExpiresAt: &end}
	require.NoError(t, subs.Create(ctx, extra))
	require.NoError(t, NewSubscriptionService(subs, repositories.NewPlanRepository(db)).Revoke(ctx, userID))
	var count int64
	require.NoError(t, db.Model(&models.Subscription{}).Where("user_id = ? AND status = ?", userID, enums.SubscriptionStatusActive).Count(&count).Error)
	require.Zero(t, count)
	// An invalid/unknown payment is different from storage failure (unit coverage
	// pins outage ack 99; this query exercises the real not-found path).
	unknown := strings.Clone("unknown-" + uuid.NewString())
	params["vnp_TxnRef"] = unknown
	query.Set("vnp_TxnRef", unknown)
	query.Set("vnp_SecureHash", vnpaySign(params, cfg.HashSecret))
	code, err := ipn.VerifyAndHandleIPN(ctx, query)
	require.NoError(t, err)
	require.Equal(t, vnpayRspOrderNotFound, code)
}
