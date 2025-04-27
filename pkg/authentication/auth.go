package authentication

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	tokenCache   = make(map[string]cachedToken)
	cacheMutex   sync.RWMutex
	cacheTimeout = 15 * time.Minute
)

type cachedToken struct {
	ValidUntil time.Time
	Claims     map[string]interface{}
}

// ValidateRequest validates Authorization: Bearer token from context
func ValidateRequest(c *gin.Context) (map[string]interface{}, error) {
	log.Print("are we here to validate token?")
	token, err := extractBearerToken(c)
	if err != nil {
		return nil, err
	}

	claims, err := validateToken(token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func extractBearerToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("authorization header format must be Bearer {token}")
	}
	return parts[1], nil
}

func validateToken(token string) (map[string]interface{}, error) {
	cacheMutex.RLock()
	//cached, found := tokenCache[token]
	cacheMutex.RUnlock()

	//if found && time.Now().Before(cached.ValidUntil) {
	//	return cached.Claims, nil
	//}

	claims, err := introspectToken(token)
	if err != nil {
		return nil, err
	}

	active, ok := claims["active"].(bool)
	if !ok || !active {
		return nil, errors.New("token is not active")
	}

	audiences, ok := claims["aud"].([]interface{})
	if !ok {
		return nil, errors.New("invalid audience claim")
	}

	hasCDS := false
	for _, aud := range audiences {
		if audStr, ok := aud.(string); ok && audStr == "iam-cds" {
			hasCDS = true
			break
		}
	}
	if !hasCDS {
		return nil, errors.New("audience iam-cds missing")
	}

	exp := time.Now().Add(cacheTimeout)
	if expTime, ok := claims["exp"].(float64); ok {
		expFromToken := time.Unix(int64(expTime), 0)
		if expFromToken.Before(exp) {
			exp = expFromToken
		}
	}

	cacheMutex.Lock()
	tokenCache[token] = cachedToken{
		ValidUntil: exp,
		Claims:     claims,
	}
	cacheMutex.Unlock()

	return claims, nil
}

func introspectToken(token string) (map[string]interface{}, error) {
	introspectionURL := "https://localhost:9443/oauth2/introspect"
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequest("POST", introspectionURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
