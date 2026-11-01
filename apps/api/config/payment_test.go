package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPaymentConfig(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		t.Setenv("VNPAY_TMN_CODE", "TMN01")
		t.Setenv("VNPAY_HASH_SECRET", "s3cret")
		t.Setenv("VNPAY_PAY_URL", "")
		t.Setenv("VNPAY_RETURN_URL", "")

		cfg := LoadPaymentConfig()

		assert.Equal(t, "TMN01", cfg.VNPayTmnCode)
		assert.Equal(t, "s3cret", cfg.VNPayHashSecret)
		assert.Equal(t, "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html", cfg.VNPayPayURL)
		assert.Equal(t, "http://localhost:3000/payments/vnpay/return", cfg.VNPayReturnURL)
	})

	t.Run("overrides", func(t *testing.T) {
		t.Setenv("VNPAY_TMN_CODE", "TMN01")
		t.Setenv("VNPAY_HASH_SECRET", "s3cret")
		t.Setenv("VNPAY_PAY_URL", "https://pay.vnpay.vn/vpcpay.html")
		t.Setenv("VNPAY_RETURN_URL", "https://app.example.com/return")

		cfg := LoadPaymentConfig()

		assert.Equal(t, "https://pay.vnpay.vn/vpcpay.html", cfg.VNPayPayURL)
		assert.Equal(t, "https://app.example.com/return", cfg.VNPayReturnURL)
	})

	t.Run("panics when hash secret is unset", func(t *testing.T) {
		t.Setenv("VNPAY_TMN_CODE", "TMN01")
		t.Setenv("VNPAY_HASH_SECRET", "")

		require.Panics(t, func() { LoadPaymentConfig() })
	})
}
