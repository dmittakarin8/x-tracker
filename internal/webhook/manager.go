package webhook

import (
    "x-tracker/internal/api"
    "x-tracker/internal/db"
    "x-tracker/internal/logger"
)

type NotificationManager struct {
    discord  *DiscordWebhook
    telegram *TelegramWebhook
    config   struct {
        enableDiscord  bool
        enableTelegram bool
    }
}

func NewNotificationManager(discordURL, telegramToken, telegramChatID string, enableDiscord, enableTelegram bool) *NotificationManager {
    manager := &NotificationManager{
        config: struct {
            enableDiscord  bool
            enableTelegram bool
        }{
            enableDiscord:  enableDiscord,
            enableTelegram: enableTelegram,
        },
    }
    
    if enableDiscord && discordURL != "" {
        manager.discord = NewDiscordWebhook(discordURL)
    }
    
    if enableTelegram && telegramToken != "" && telegramChatID != "" {
        manager.telegram = NewTelegramWebhook(telegramToken, telegramChatID)
    }
    
    return manager
}

func (m *NotificationManager) NotifyNewFollows(account *db.WatchedAccount, follows []string, api *api.Client) {
    if m.config.enableDiscord && m.discord != nil {
        if err := m.discord.NotifyNewFollows(account, follows, api); err != nil {
            logger.Info("Failed to send Discord follow notification: %v", err)
        }
    }
    
    if m.config.enableTelegram && m.telegram != nil {
        if err := m.telegram.NotifyNewFollows(account, follows, api); err != nil {
            logger.Info("Failed to send Telegram follow notification: %v", err)
        }
    }
}

func (m *NotificationManager) NotifyUnfollows(account *db.WatchedAccount, unfollows []string, api *api.Client) {
    if m.config.enableDiscord && m.discord != nil {
        if err := m.discord.NotifyUnfollows(account, unfollows, api); err != nil {
            logger.Info("Failed to send Discord unfollow notification: %v", err)
        }
    }
    
    if m.config.enableTelegram && m.telegram != nil {
        if err := m.telegram.NotifyUnfollows(account, unfollows, api); err != nil {
            logger.Info("Failed to send Telegram unfollow notification: %v", err)
        }
    }
} 