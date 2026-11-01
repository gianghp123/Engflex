package config

import (
	"engflex-api/internal/utils"
)

// PaymentConfig holds the VNPay gateway settings. TmnCode and HashSecret
// are required so the server fails fast when unset — a missing secret must
// never silently fall back to an empty-string signature.
type PaymentConfig struct {
	// VNPayTmnCode is the merchant terminal code VNPay assigned.
	VNPayTmnCode string
	// VNPayHashSecret signs/verifies every checkout URL and IPN callback.
	VNPayHashSecret string
	// VNPayPayURL is the gateway's checkout endpoint (sandbox by default).
	VNPayPayURL string
	// VNPayReturnURL is where VNPay redirects the learner's browser after
	// payment — UX-only, never trusted alone to grant access.
	VNPayReturnURL string
}

// LoadPaymentConfig reads the payment configuration from the environment.
func LoadPaymentConfig() PaymentConfig {
	return PaymentConfig{
		VNPayTmnCode:    utils.MustGetEnv("VNPAY_TMN_CODE"),
		VNPayHashSecret: utils.MustGetEnv("VNPAY_HASH_SECRET"),
		VNPayPayURL:     utils.GetEnv("VNPAY_PAY_URL", "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"),
		VNPayReturnURL:  utils.GetEnv("VNPAY_RETURN_URL", "http://localhost:3000/payments/vnpay/return"),
	}
}
