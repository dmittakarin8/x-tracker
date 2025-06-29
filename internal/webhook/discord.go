package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"x-tracker/internal/db"
	"x-tracker/internal/logger"
	"x-tracker/internal/api"
)

type DiscordWebhook struct {
	URL        string
	httpClient *http.Client
}

type webhookPayload struct {
	Username  string         `json:"username"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []webhookEmbed `json:"embeds"`
}

type webhookEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Fields      []webhookEmbedField `json:"fields"`
	Timestamp   string              `json:"timestamp"`
	Footer      webhookEmbedFooter  `json:"footer"`
}

type webhookEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type webhookEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

func NewDiscordWebhook(webhookURL string) *DiscordWebhook {
	return &DiscordWebhook{
		URL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *DiscordWebhook) send(payload webhookPayload) error {
	// Add logging for webhook URL
	logger.Info("Attempting to send Discord webhook to URL: %s", d.URL)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling webhook payload: %w", err)
	}

	// Log the payload being sent
	logger.Info("Sending webhook payload: %s", string(jsonData))

	resp, err := d.httpClient.Post(d.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	// Log the response status
	logger.Info("Discord webhook response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("webhook error: status=%d", resp.StatusCode)
	}

	logger.Info("Successfully sent Discord webhook notification")
	return nil
}

func (d *DiscordWebhook) NotifyNewFollows(account *db.WatchedAccount, follows []string, api *api.Client) error {
	if d.URL == "" {
		logger.Info("Discord webhook URL is empty, skipping follow notification")
		return nil
	}

	logger.Info("Preparing follow notification for %s: +%d follows", account.Username, len(follows))

	followEmbed := webhookEmbed{
		Title:       fmt.Sprintf("New Follows Detected for @%s", account.Username),
		Description: fmt.Sprintf("Started following %d new accounts", len(follows)),
		Color:       0x00ff00,
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields:      make([]webhookEmbedField, 0, len(follows)),
		Footer: webhookEmbedFooter{
			Text: "X Track",
		},
	}

	// Add fields for each new follow (up to 25)
	for i, userID := range follows {
		if i < 25 {
			userDetails, err := api.GetUserByID(userID)
			var username string
			var following_followers int
			if err != nil {
				logger.Info("Failed to get username for ID %s: %v", userID, err)
				username = userID
			} else {
				username = fmt.Sprintf("@%s", userDetails.Legacy.ScreenName)
				following_followers = userDetails.Legacy.FollowersCount
			}

			followEmbed.Fields = append(followEmbed.Fields, webhookEmbedField{
				Name:   fmt.Sprintf("New Follow %d", i+1),
				Value:  username + " " + fmt.Sprintf("%d followers", following_followers),
				Inline: true,
			})
		}
	}

	payload := webhookPayload{
		Username: "X Follow Tracker",
		Embeds:   []webhookEmbed{followEmbed},
	}

	return d.send(payload)
}

func (d *DiscordWebhook) NotifyUnfollows(account *db.WatchedAccount, unfollows []string, api *api.Client) error {
	if d.URL == "" {
		logger.Info("Discord webhook URL is empty, skipping unfollow notification")
		return nil
	}

	logger.Info("Preparing unfollow notification for %s: -%d unfollows", account.Username, len(unfollows))

	unfollowEmbed := webhookEmbed{
		Title:       fmt.Sprintf("Unfollows Detected for @%s", account.Username),
		Description: fmt.Sprintf("Unfollowed %d accounts", len(unfollows)),
		Color:       0xFF0000,
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields:      make([]webhookEmbedField, 0, len(unfollows)),
		Footer: webhookEmbedFooter{
			Text: "X Track",
		},
	}

	// Add fields for each unfollow (up to 25)
	for i, userID := range unfollows {
		if i < 25 {
			userDetails, err := api.GetUserByID(userID)
			var username string
			var following_followers int
			if err != nil {
				logger.Info("Failed to get username for ID %s: %v", userID, err)
				username = userID
			} else {
				username = fmt.Sprintf("@%s", userDetails.Legacy.ScreenName)
				following_followers = userDetails.Legacy.FollowersCount
			}

			unfollowEmbed.Fields = append(unfollowEmbed.Fields, webhookEmbedField{
				Name:   fmt.Sprintf("Unfollow %d", i+1),
				Value:  username + " " + fmt.Sprintf("%d followers", following_followers),
				Inline: true,
			})
		}
	}

	payload := webhookPayload{
		Username: "X Follow Tracker",
		Embeds:   []webhookEmbed{unfollowEmbed},
	}

	return d.send(payload)
}

func (d *DiscordWebhook) NotifyFollowingChange(username string, newCount int) error {
	if d.URL == "" {
		return nil // Webhook notifications disabled
	}

	embed := webhookEmbed{
		Title:       fmt.Sprintf("Following Count Changed for @%s", username),
		Description: fmt.Sprintf("New following count: %d", newCount),
		Color:       0xFFA500, // Orange for changes
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: webhookEmbedFooter{
			Text: "CLI X Track",
		},
	}

	payload := webhookPayload{
		Username: "X Follow Tracker",
		Embeds:   []webhookEmbed{embed},
	}

	return d.send(payload)
} 