package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	utils "social-network/shared/go/http-utils"
	"strings"
	"time"
)

// ======== Minimal custom JWT (HS256) using ONLY Go stdlib ========
// This implementation creates and validates compact JWTs (header.payload.signature)
// with HMAC-SHA256. It keeps things small and readable for learning/testing.
//
// Security notes:
// - Use a long, random secret from env/secret manager in production.
// - Keep token lifetimes short and rotate secrets periodically.
// - Consider adding jti (token id) and a store if you need revocation.
// - Clock skew tolerance is applied for nbf/exp checks.

// Claims represents the token payload. Add fields you need.
type Claims struct {
	UserId int64 `json:"user_id"`       // user ID
	Exp    int64 `json:"exp"`           // expiration (unix seconds)
	Iat    int64 `json:"iat"`           // issued-at (unix seconds)
	Nbf    int64 `json:"nbf,omitempty"` // not-before (unix seconds)
	// Roles  []string `json:"roles,omitempty"`
	// You can embed custom fields as needed, e.g. Email, TenantID, etc.
}

var (
	secret = []byte(os.Getenv("JWT_KEY"))
	// Allow a small clock skew when validating nbf/exp.
	clockSkew = 30 * time.Second
)

// CreateToken builds a signed JWT string with HS256.
func CreateToken(claims Claims) (string, error) {
	head := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(head)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEnc := utils.B64urlEncode(headerJSON)
	payloadEnc := utils.B64urlEncode(payloadJSON)
	unsigned := headerEnc + "." + payloadEnc

	sig := signHS256(unsigned, secret)
	return unsigned + "." + sig, nil
}

func signHS256(unsigned string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(unsigned))
	return utils.B64urlEncode(h.Sum(nil))
}

// ParseAndValidate verifies the signature and time-based claims.
func ParseAndValidate(token string) (Claims, error) {
	var zero Claims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return zero, errors.New("invalid token format")
	}
	unsigned := parts[0] + "." + parts[1]
	expected := signHS256(unsigned, secret)
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return zero, errors.New("invalid signature")
	}

	payload, err := utils.B64urlDecode(parts[1])
	if err != nil {
		return zero, fmt.Errorf("payload base64: %w", err)
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return zero, fmt.Errorf("payload json: %w", err)
	}

	now := time.Now()
	if claims.Nbf != 0 {
		if now.Add(clockSkew).Before(time.Unix(claims.Nbf, 0)) {
			return zero, errors.New("token not yet valid")
		}
	}
	if claims.Exp != 0 {
		if now.After(time.Unix(claims.Exp, 0).Add(clockSkew)) {
			return zero, errors.New("token expired")
		}
	}
	return claims, nil
}
