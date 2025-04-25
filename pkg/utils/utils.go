package utils

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wso2/identity-customer-data-service/pkg/constants"
	"github.com/wso2/identity-customer-data-service/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	traceID := c.GetString("traceId")

	switch e := err.(type) {
	case *errors.ClientError:
		c.JSON(e.StatusCode, gin.H{
			"error_code":        e.Code,
			"error_message":     e.Message,
			"error_description": e.Description,
			"traceId":           traceID,
		})
	case *errors.ServerError:
		log.Printf("[ERROR] %s | code: %s | message: %s | traceId: %s\n", e.Description, e.Code, e.Message, traceID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_code":        e.Code,
			"error_message":     e.Message,
			"error_description": e.Description,
			"traceId":           traceID,
		})
	default:
		log.Printf("[ERROR] Unknown error: %v | traceId: %s\n", err, traceID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_code":        "50000",
			"error_message":     "Internal Server Error",
			"error_description": "An unexpected error occurred.",
			"traceId":           traceID,
		})
	}
}

func NormalizePropertyType(propertyType string) (string, error) {
	normalized := strings.ToLower(propertyType)
	if goType, ok := constants.GoTypeMapping[normalized]; ok {
		return goType, nil
	}
	return "", fmt.Errorf("unsupported property type: %s", propertyType)
}

// MergeStringValues merges two string values based on the strategy.
func MergeStringValues(existing string, incoming string, strategy string) string {
	switch strategy {
	case "overwrite":
		return incoming
	case "ignore":
		return existing
	default: // default to "combine"
		if existing == "" {
			return incoming
		}
		if incoming == "" || existing == incoming {
			return existing
		}
		return existing + " | " + incoming
	}
}

// MergeStringSlices merges two string slices based on the strategy.
func MergeStringSlices(existing []string, incoming []string, strategy string) []string {
	switch strategy {
	case "overwrite":
		return incoming
	case "ignore":
		return existing
	default: // default to "combine"
		unique := map[string]bool{}
		for _, v := range existing {
			unique[v] = true
		}
		for _, v := range incoming {
			unique[v] = true
		}
		var merged []string
		for val := range unique {
			merged = append(merged, val)
		}
		return merged
	}
}

// ApplyMasking applies the given masking strategy to a string value.
func ApplyMasking(value string, strategy string) string {
	switch strings.ToLower(strategy) {
	case "partial":
		return maskPartial(value)
	case "hash":
		return hashValue(value)
	case "redact":
		return "REDACTED"
	default:
		return value // no masking
	}
}

// maskPartial masks the middle part of a string (e.g., email)
func maskPartial(value string) string {
	if len(value) <= 4 {
		return "***"
	}
	visible := 2
	masked := strings.Repeat("*", len(value)-2*visible)
	return value[:visible] + masked + value[len(value)-visible:]
}

// hashValue returns a SHA256 hash of the value
func hashValue(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

// ReverseMasking returns the visible portions of a partially masked string.
func ReverseMasking(maskedValue, strategy string) string {
	switch strings.ToLower(strategy) {
	case "partial":
		return getVisibleFromPartial(maskedValue)
	default:
		return "" // not reversible
	}
}

// getVisibleFromPartial extracts the first and last 2 characters
func getVisibleFromPartial(value string) string {
	if len(value) <= 4 {
		return ""
	}
	return value[:2] + "..." + value[len(value)-2:]
}

// GetUserDataFromSCIM fetches user details from the SCIM2 endpoint using Bearer token
func GetUserDataFromSCIM(token string, userId string) (map[string]interface{}, error) {
	// ⚠️ Disable TLS verification for dev
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	url := fmt.Sprintf("https://localhost:9443/scim2/Users/%s", userId)
	log.Print("Fetching user data from SCIM for url===", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create SCIM request: %v", err)
	}

	// Set headers
	auth := "admin:admin"
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))

	// Step 2: Add Authorization header
	log.Println("encoded===", encoded)
	req.Header.Add("Authorization", "Basic "+encoded)
	req.Header.Add("Accept", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("status code==", err.Error())
		return nil, fmt.Errorf("error sending request to SCIM endpoint: %v", err)
	}
	defer resp.Body.Close()

	log.Println("status code==", resp.StatusCode)
	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("SCIM request failed: status %d - %s", resp.StatusCode, string(body))
	}

	// Decode response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode SCIM response: %v", err)
	}
	log.Print("SCIM response: ", result)
	return result, nil
}

func ExtractIdentityFromSCIM(scim map[string]interface{}) map[string]interface{} {
	identity := make(map[string]interface{})

	// Step 1: Handle special known fields with key renaming
	if userName, ok := scim["userName"]; ok {
		identity["user_name"] = userName
	}
	if userId, ok := scim["id"]; ok {
		identity["user_id"] = userId
	}

	// Step 2: Preserve "name", "emails", "roles", "groups" if present
	if name, ok := scim["name"]; ok {
		identity["name"] = name
	}
	if emails, ok := scim["emails"]; ok {
		identity["emails"] = emails
	}
	if rawRoles, ok := scim["roles"]; ok {
		if rolesArray, ok := rawRoles.([]interface{}); ok {
			var simplifiedRoles []map[string]interface{}
			for _, role := range rolesArray {
				if roleMap, ok := role.(map[string]interface{}); ok {
					simplifiedRole := map[string]interface{}{}
					if val, ok := roleMap["audienceType"]; ok {
						simplifiedRole["audience_type"] = val
					}
					if val, ok := roleMap["display"]; ok {
						simplifiedRole["display"] = val
					}
					simplifiedRoles = append(simplifiedRoles, simplifiedRole)
				}
			}
			identity["roles"] = simplifiedRoles
		}
	}

	if groups, ok := scim["groups"]; ok {
		identity["groups"] = groups
	}

	// Step 3: Flatten all SCIM schema extensions like `urn:*`
	for key, value := range scim {
		if strings.HasPrefix(key, "urn:") {
			if nestedMap, ok := value.(map[string]interface{}); ok {
				for k, v := range nestedMap {
					identity[k] = v // flatten into top-level identity
				}
			}
		}
	}

	// Optional: Remove noisy SCIM fields you don’t care about
	delete(identity, "schemas")
	delete(identity, "meta")

	return identity
}
