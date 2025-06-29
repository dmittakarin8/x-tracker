package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"x-tracker/config"
	"x-tracker/internal/api"
	"x-tracker/internal/db"
	"x-tracker/internal/ui"
	"x-tracker/internal/webhook"
	"x-tracker/internal/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize logger
	if err := logger.Initialize(cfg.LoggingEnabled, cfg.LogDir); err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Close()

	logger.Info("CLI X Track starting up...")

	// Initialize database
	database, err := db.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer database.Close()

	// Initialize API client
	apiClient := api.NewClient(cfg)

	// Initialize notification manager
	notificationManager := webhook.NewNotificationManager(
		cfg.DiscordWebhookURL,
		cfg.TelegramBotToken,
		cfg.TelegramChatID,
		cfg.EnableDiscordNotifications,
		cfg.EnableTelegramNotifications,
	)

	// Initialize UI model with notification manager
	model := ui.NewModel(database, apiClient, notificationManager, cfg)

	// Create and start the Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		p.Kill()
	}()

	// Run the application
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
