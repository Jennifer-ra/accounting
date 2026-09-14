package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	UserID uint  `json:"uid"`
	Exp    int64 `json:"exp"`
}

func Sign(secret string, userID uint, ttl time.Duration) (string, error) {
	claims := Claims{UserID: userID, Exp: time.Now().Add(ttl).Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadText := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(secret, payloadText)
	return payloadText + "." + sig, nil
}

func Parse(secret, token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, errors.New("invalid token")
	}
	if !hmac.Equal([]byte(parts[1]), []byte(sign(secret, parts[0]))) {
		return Claims{}, errors.New("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, err
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, err
	}
	if claims.UserID == 0 || claims.Exp < time.Now().Unix() {
		return Claims{}, errors.New("token expired")
	}
	return claims, nil
}

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	mac.Write([]byte("." + strconv.FormatInt(int64(len(payload)), 10)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func Bearer(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("missing bearer token")
	}
	return strings.TrimSpace(parts[1]), nil
}
