package services

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// vnpaySign computes VNPay's HMAC-SHA512 signature over params: sort keys,
// join as "key1=value1&key2=value2..." with url.QueryEscape per VNPay's own
// escaping (space -> "+", matching PHP's urlencode, which net/url.QueryEscape
// also produces), then HMAC-SHA512 the joined string with secret.
func vnpaySign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(url.QueryEscape(k))
		sb.WriteByte('=')
		sb.WriteString(url.QueryEscape(params[k]))
	}

	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(sb.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

// vnpaySignQuery is vnpaySign over an already-parsed query (used to verify
// an inbound IPN/return callback, which arrives as url.Values with
// vnp_SecureHash itself excluded from the signed set).
func vnpaySignQuery(query url.Values, secret string) string {
	params := make(map[string]string, len(query))
	for k, vals := range query {
		if k == "vnp_SecureHash" || k == "vnp_SecureHashType" || len(vals) == 0 {
			continue
		}
		params[k] = vals[0]
	}
	return vnpaySign(params, secret)
}
