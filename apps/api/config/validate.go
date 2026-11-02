package config

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Validate names invalid settings without echoing credentials or provider URLs.
func (c Config) Validate() error {
	var failures []error
	required := map[string]string{"CLERK_SECRET_KEY": c.Auth.ClerkSecretKey,
		"LLM_API_KEY": c.LLM.APIKey, "VOICE_INTERNAL_SECRET": c.Voice.InternalSecret,
		"VNPAY_TMN_CODE": c.Payment.VNPayTmnCode, "VNPAY_HASH_SECRET": c.Payment.VNPayHashSecret}
	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			failures = append(failures, fmt.Errorf("%s is required", key))
		}
	}
	port, err := strconv.Atoi(c.Server.Port)
	if err != nil || port < 1 || port > 65535 {
		failures = append(failures, errors.New("PORT must be between 1 and 65535"))
	}
	endpoints := map[string]string{"VOICE_SERVICE_URL": c.Voice.ServiceURL, "PRONOUNCE_SERVICE_URL": c.Voice.PronounceURL, "LLM_BASE_URL": c.LLM.BaseURL, "VNPAY_PAY_URL": c.Payment.VNPayPayURL, "VNPAY_RETURN_URL": c.Payment.VNPayReturnURL}
	for key, value := range endpoints {
		u, err := url.Parse(value)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			failures = append(failures, fmt.Errorf("%s must be an absolute HTTP(S) URL", key))
		}
	}
	if c.Database.MaxOpenConns < 1 || c.Database.MaxIdleConns < 0 || c.Database.MaxIdleConns > c.Database.MaxOpenConns || c.Database.ConnMaxLifetime <= 0 || c.Database.ConnMaxIdleTime <= 0 {
		failures = append(failures, errors.New("invalid database pool limits"))
	}
	if c.Voice.StartTimeout <= 0 || c.Voice.OfferTimeout <= 0 || c.Voice.AnalyzeTimeout <= 0 || c.Voice.PronounceTimeout <= 0 || c.LLM.Timeout <= 0 {
		failures = append(failures, errors.New("provider timeouts must be positive"))
	}
	return errors.Join(failures...)
}
