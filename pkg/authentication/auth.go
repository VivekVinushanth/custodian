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

// ValidateAuthentication validates Authorization: Bearer token from context
func ValidateAuthentication(c *gin.Context) (map[string]interface{}, error) {
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

func GetTokenFromIS(applicationID string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Only for local/dev
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	endpoint := "https://localhost:9443/oauth2/token"

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	//data.Set("scope", "test")
	data.Set("tokenBindingId", applicationID) // Add application ID as token binding ID

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	// Basic Auth Header (e.g., client_id:client_secret)
	auth := "k06eyXqdJvoSBx_steWLWdCruBca:fjxoAQVCJlvTprKxZVd3tIl733fzWrvB5gJcKgqBBRYa" // Replace with actual client credentials if available
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))

	req.Header.Add("Authorization", "Basic "+encoded)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending token request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed: status %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in response")
	}

	log.Printf("New access token generated: %s", accessToken)
	return accessToken, nil
}

func RevokeToken(token string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	endpoint := "https://localhost:9443/oauth2/revoke"

	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	// Basic Auth Header (same as token endpoint)
	auth := "k06eyXqdJvoSBx_steWLWdCruBca:fjxoAQVCJlvTprKxZVd3tIl733fzWrvB5gJcKgqBBRYa" // Replace with actual client credentials if available
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))

	req.Header.Add("Authorization", "Basic "+encoded)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending revoke request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke request failed: status %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("Token revoked successfully: %s", token)
	return nil
}
