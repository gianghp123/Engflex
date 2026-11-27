package config

import "engflex-api/internal/utils"

// SepayConfig holds the SePay gateway settings. APIKey and WebhookSecret are
// unrelated credentials from two different SePay dashboard pages: APIKey is
// sent as a Bearer token when this server calls SePay's own API (fetching
// the merchant's bank account for QR generation); WebhookSecret is the
// arbitrary secret chosen when creating the webhook under "Tích hợp
// WebHooks", compared against the incoming "Authorization: Apikey ..."
// header to verify a payment notification really came from SePay.
type SepayConfig struct {
	APIKey        string
	WebhookSecret string
	BaseURL       string

	// Bank account the QR codes pay into. Set these to skip the API
	// lookup entirely — the account changes about never, and an API key
	// issued for a different SePay organisation than the linked bank
	// account returns the wrong account rather than an error, which is a
	// silent mismatch that is hard to spot. APIKey is only needed when
	// these are left blank.
	BankAccount   string
	BankName      string
	AccountHolder string
}

// LoadSepayConfig reads the SePay configuration from the environment.
func LoadSepayConfig() SepayConfig {
	return SepayConfig{
		APIKey:        utils.GetEnv("SEPAY_API_KEY", ""),
		WebhookSecret: utils.MustGetEnv("SEPAY_WEBHOOK_SECRET"),
		BaseURL:       utils.GetEnv("SEPAY_API_BASE_URL", "https://userapi.sepay.vn/v2"),
		BankAccount:   utils.GetEnv("SEPAY_BANK_ACCOUNT", ""),
		BankName:      utils.GetEnv("SEPAY_BANK_NAME", ""),
		AccountHolder: utils.GetEnv("SEPAY_ACCOUNT_HOLDER", ""),
	}
}
