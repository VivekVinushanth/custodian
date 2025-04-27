package handlers

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wso2/identity-customer-data-service/pkg/authentication"
	"github.com/wso2/identity-customer-data-service/pkg/models"
	"github.com/wso2/identity-customer-data-service/pkg/service"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// AddEvent handles adding a single event
func (s Server) AddEvent(c *gin.Context) {

	// Step 1: Validate token
	_, err := authentication.ValidateRequest(c)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := service.AddEvents(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Event added successfully"})
}

// GetUserEvent fetches a specific event
func (s Server) GetEvent(c *gin.Context, eventId string) {
	//TODO implement me
	panic("implement me")
}

// TODO remove
func (s Server) GetEvents(c *gin.Context) {
	// Step 1: Extract raw filters (e.g., event_type+eq+Identify)
	rawFilters := c.QueryArray("filter")
	log.Println("Filters11: ", rawFilters)

	// Step 2: Parse optional time range
	var timeFilter bson.M
	if timeStr := c.Query("time_range"); timeStr != "" {
		log.Println("Time Rangedfff: ", timeStr)
		durationSec, _ := strconv.Atoi(timeStr)       // parse string to int
		currentTime := time.Now().UTC().Unix()        // current time in seconds
		startTime := currentTime - int64(durationSec) // assuming value is in minutes
		timeFilter = bson.M{
			"event_timestamp": bson.M{"$gte": startTime},
		}
	}

	// Step 3: Fetch events with filter strings
	events, err := service.GetEvents(rawFilters, timeFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

func (s Server) GetWriteKey(c *gin.Context, application_id string) {
	// Step 1: Get existing token if needed (for now assume no previous token available)
	// If you have a DB or cache, fetch the existing token here.
	existingToken, _ := GetTokenFromIS(application_id)

	// Step 2: If token exists, revoke it first
	if existingToken != "" {
		log.Println("revoking existing as it is considereed re-gerneate")
		err := RevokeToken(existingToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke existing token", "details": err.Error()})
			return
		}
	}

	// Step 3: Get a new token
	newToken, err := GetTokenFromIS(application_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch new token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"write_key": newToken})
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
