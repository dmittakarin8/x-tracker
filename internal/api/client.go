package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"x-tracker/config"
	"x-tracker/internal/logger"
)

type Client struct {
	httpClient *http.Client
	config     *config.Config
	remainingRequests int32  // Using atomic for thread safety
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		config: cfg,
	}
}

func (c *Client) GetUser(username string) (*UserResponse, error) {
	logger.Info("Starting user lookup for: %s", username)
	
	url := fmt.Sprintf("https://%s/v2/user/by-username?username=%s", c.config.RapidAPIHost, username)
	logger.Info("Making request to: %s", url)
	
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var response UserResponse
	if err := c.doRequest(req, &response); err != nil {
		logger.Info("User lookup failed for %s: %v", username, err)
		return nil, err
	}

	logger.Info("User lookup completed for %s (ID: %s) with a following count of %d", 
		username, response.RestID, response.Legacy.FriendsCount)
	return &response, nil
}

func (c *Client) GetFollowingIDs(userID string) (*FollowingIDsResponse, error) {
	var allIDs []string
	nextCursor := "0"
	
	for {
		endpoint := fmt.Sprintf("https://%s/v2/user/following-ids", c.config.RapidAPIHost)
		
		// Build query parameters
		params := url.Values{}
		params.Add("userId", userID)
		params.Add("count", "5000")
		if nextCursor != "0" {
			params.Add("cursor", nextCursor)
		}
		
		req, err := c.newRequest("GET", endpoint+"?"+params.Encode(), nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		var response FollowingIDsResponse
		if err := c.doRequest(req, &response); err != nil {
			return nil, fmt.Errorf("sending request: %w", err)
		}

		// Append the current page of IDs
		allIDs = append(allIDs, response.IDs...)

		// Check if we need to fetch more pages
		if response.NextCursor == 0 {
			break
		}
		nextCursor = response.NextCursorStr

		// Add a small delay to avoid rate limiting
		time.Sleep(time.Second)
		
		logger.Info("client.go.GetFollowingIDs - Fetching next page with cursor: %s", nextCursor)
	}
    logger.Info("client.go.GetFollowingIDs - Fetched a total of %d IDs for user %s", len(allIDs), userID)
	// Return all collected IDs in the response structure
	return &FollowingIDsResponse{
		IDs: allIDs,
	}, nil
}

func (c *Client) GetUserByID(userID string) (*UserByIDResponse, error) {
	logger.Info("Looking up user by ID: %s", userID)
	
	url := fmt.Sprintf("https://%s/v2/user/by-id?userId=%s", 
		c.config.RapidAPIHost, userID)
	
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var response UserByIDResponse
	if err := c.doRequest(req, &response); err != nil {
		logger.Info("User lookup failed for ID %s: %v", userID, err)
		return nil, err
	}

	logger.Info("User lookup completed for ID %s: @%s with %d followers", userID, response.Legacy.ScreenName, response.Legacy.FollowersCount)
	return &response, nil
}

func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-rapidapi-key", c.config.RapidAPIKey)
	req.Header.Add("x-rapidapi-host", c.config.RapidAPIHost)

	logger.Info("Request headers: Host=%s", c.config.RapidAPIHost)

	return req, nil
}

func (c *Client) doRequest(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	// Check rate limit header
	if remaining := resp.Header.Get("x-ratelimit-requests-remaining"); remaining != "" {
		if count, err := strconv.Atoi(remaining); err == nil {
			atomic.StoreInt32(&c.remainingRequests, int32(count))
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}

// Add getter for remaining requests
func (c *Client) RemainingRequests() int {
	return int(atomic.LoadInt32(&c.remainingRequests))
} 