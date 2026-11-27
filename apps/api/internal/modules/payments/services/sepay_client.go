package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SepayBankAccount is the subset of SePay's bank-account record this app
// needs to render a VietQR image — see GET {BaseURL}/bank-accounts.
type SepayBankAccount struct {
	AccountNumber     string `json:"account_number"`
	AccountHolderName string `json:"account_holder_name"`
	BankShortName     string `json:"bank_short_name"`
}

type sepayBankAccountsResponse struct {
	Status string             `json:"status"`
	Data   []SepayBankAccount `json:"data"`
}

// sepayClient calls SePay's own API (not the webhook it sends us — that's
// verified separately in sepay_service.go). Used once per checkout to
// resolve which bank account to build a QR against; the result is cached
// in-memory since a merchant's linked accounts change essentially never.
type sepayClient struct {
	baseURL string
	apiKey  string
	http    *http.Client

	mu       sync.Mutex
	cached   *SepayBankAccount
	cachedAt time.Time
}

const sepayBankAccountCacheTTL = 10 * time.Minute

func newSepayClient(baseURL, apiKey string) *sepayClient {
	return &sepayClient{baseURL: baseURL, apiKey: apiKey, http: &http.Client{Timeout: 10 * time.Second}}
}

// PrimaryBankAccount returns the merchant's first active linked bank
// account. SePay accounts typically link just one account for receiving
// payments; if several are linked, the first one SePay returns wins —
// there's no per-plan routing to choose between them.
func (c *sepayClient) PrimaryBankAccount(ctx context.Context) (*SepayBankAccount, error) {
	c.mu.Lock()
	if c.cached != nil && time.Since(c.cachedAt) < sepayBankAccountCacheTTL {
		cached := *c.cached
		c.mu.Unlock()
		return &cached, nil
	}
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/bank-accounts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sepay bank-accounts: unexpected status %d", resp.StatusCode)
	}
	var parsed sepayBankAccountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("sepay bank-accounts: no linked account")
	}
	account := parsed.Data[0]

	c.mu.Lock()
	c.cached = &account
	c.cachedAt = time.Now()
	c.mu.Unlock()
	return &account, nil
}
