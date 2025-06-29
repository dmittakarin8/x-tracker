package webhook

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"
    
    "x-tracker/internal/api"
    "x-tracker/internal/db"
    "x-tracker/internal/logger"
)

type TelegramWebhook struct {
    botToken string
    chatID   string
    client   *http.Client
}

func NewTelegramWebhook(botToken, chatID string) *TelegramWebhook {
    return &TelegramWebhook{
        botToken: botToken,
        chatID:   chatID,
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (t *TelegramWebhook) sendMessage(text string) error {
    if t.botToken == "" || t.chatID == "" {
        logger.Info("Telegram configuration missing, skipping notification")
        return nil
    }

    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
    
    payload := map[string]interface{}{
        "chat_id":    t.chatID,
        "text":       text,
        "parse_mode": "HTML",
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshaling telegram payload: %w", err)
    }
    
    resp, err := t.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("sending telegram message: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("telegram API error: status=%d", resp.StatusCode)
    }
    
    return nil
}

func (t *TelegramWebhook) NotifyNewFollows(account *db.WatchedAccount, follows []string, api *api.Client) error {
    var message strings.Builder
    
    fmt.Fprintf(&message, "<b>New Follows Detected for @%s</b>\n", account.Username)
    fmt.Fprintf(&message, "Started following %d new accounts\n\n", len(follows))
    
    // Add details for each new follow (up to 25)
    for i, userID := range follows {
        if i >= 25 {
            break
        }
        
        userDetails, err := api.GetUserByID(userID)
        if err != nil {
            logger.Info("Failed to get username for ID %s: %v", userID, err)
            fmt.Fprintf(&message, "%d. ID: %s\n", i+1, userID)
        } else {
            fmt.Fprintf(&message, "%d. @%s (%d followers)\n", 
                i+1, 
                userDetails.Legacy.ScreenName,
                userDetails.Legacy.FollowersCount)
        }
    }
    
    return t.sendMessage(message.String())
}

func (t *TelegramWebhook) NotifyUnfollows(account *db.WatchedAccount, unfollows []string, api *api.Client) error {
    var message strings.Builder
    
    fmt.Fprintf(&message, "<b>Unfollows Detected for @%s</b>\n", account.Username)
    fmt.Fprintf(&message, "Unfollowed %d accounts\n\n", len(unfollows))
    
    // Add details for each unfollow (up to 25)
    for i, userID := range unfollows {
        if i >= 25 {
            break
        }
        
        userDetails, err := api.GetUserByID(userID)
        if err != nil {
            logger.Info("Failed to get username for ID %s: %v", userID, err)
            fmt.Fprintf(&message, "%d. ID: %s\n", i+1, userID)
        } else {
            fmt.Fprintf(&message, "%d. @%s (%d followers)\n", 
                i+1, 
                userDetails.Legacy.ScreenName,
                userDetails.Legacy.FollowersCount)
        }
    }
    
    return t.sendMessage(message.String())
} 