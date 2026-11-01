package services

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVnpaySign_RoundTrip(t *testing.T) {
	params := map[string]string{
		"vnp_Version": "2.1.0",
		"vnp_Command": "pay",
		"vnp_TmnCode": "TMN01",
		"vnp_Amount":  "9900000",
		"vnp_TxnRef":  "ref-1",
	}
	hash := vnpaySign(params, "secret")
	assert.NotEmpty(t, hash)

	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("vnp_SecureHash", hash)

	assert.Equal(t, hash, vnpaySignQuery(q, "secret"))
}

func TestVnpaySign_TamperingIsRejected(t *testing.T) {
	params := map[string]string{
		"vnp_TmnCode": "TMN01",
		"vnp_Amount":  "9900000",
		"vnp_TxnRef":  "ref-1",
	}
	hash := vnpaySign(params, "secret")

	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("vnp_Amount", "1") // tampered after signing
	q.Set("vnp_SecureHash", hash)

	assert.NotEqual(t, hash, vnpaySignQuery(q, "secret"))
}

func TestVnpaySign_WrongSecretIsRejected(t *testing.T) {
	params := map[string]string{"vnp_TxnRef": "ref-1"}
	hash := vnpaySign(params, "secret-a")
	assert.NotEqual(t, hash, vnpaySign(params, "secret-b"))
}
