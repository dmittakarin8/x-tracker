package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"x-tracker/config"
	"x-tracker/internal/api"
	"x-tracker/internal/db"
	"x-tracker/internal/webhook"
	"x-tracker/internal/logger"
)

// Add back just the uptime tick message type
type tickMsg time.Time

// Add a new message type for check timer
type checkTimerMsg time.Time

// Message types
type (
	errMsg error
	CheckAccountsMsg time.Time
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeAddAccount
	ModeListAccounts
	ModeRemoveAccount

	// Braille spinner characters
	brailleSpinnerFrames = `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`
)

// Add String method for better logging
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "Normal"
	case ModeAddAccount:
		return "Add"
	case ModeListAccounts:
		return "List"
	case ModeRemoveAccount:
		return "Remove"
	default:
		return "Unknown"
	}
}

type Model struct {
	mode           Mode
	db             *db.Database
	api            *api.Client
	notifications  *webhook.NotificationManager
	config         *config.Config
	accounts       []db.WatchedAccount
	spinner        spinner.Model
	brailleSpinner spinner.Model
	error          error
	input          string
	selected       int
	uptime         time.Duration
	startTime      time.Time
	textInput      textinput.Model
	lastCheckTime  time.Time
	checkInterval  time.Duration
	lastTick       time.Time
}

func NewModel(database *db.Database, apiClient *api.Client, notifications *webhook.NotificationManager, cfg *config.Config) *Model {
	// Initialize text input with styling
	ti := textinput.New()
	ti.Placeholder = "username (without @)"
	ti.PlaceholderStyle = placeholderStyle
	ti.PromptStyle = inputPromptStyle
	ti.TextStyle = inputStyle
	ti.Cursor.Style = cursorStyle
	ti.CharLimit = 50
	ti.Width = 30
	ti.Prompt = "@ "

	// Initialize spinners with proper timing
	s := spinner.New(
		spinner.WithSpinner(spinner.Spinner{
			Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
			FPS:    time.Second / 8, // Medium speed
		}),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(special)),
	)

	bs := spinner.New(
		spinner.WithSpinner(spinner.Spinner{
			Frames: strings.Split(brailleSpinnerFrames, ""),
			FPS:    time.Second / 8, // Match the first spinner's speed
		}),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(highlight)),
	)

	return &Model{
		mode:           ModeNormal,
		db:             database,
		api:            apiClient,
		notifications: notifications,
		config:         cfg,
		spinner:        s,
		brailleSpinner: bs,
		textInput:      ti,
		startTime:      time.Now(),
		lastCheckTime:  time.Now(),
		checkInterval:  cfg.CheckInterval,
		lastTick:       time.Now(),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.brailleSpinner.Tick,
		m.tickUptime(),
		m.loadAccounts,
		m.tickCheckTimer(),
	)
}

func (m *Model) tickUptime() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Add the check timer tick function
func (m *Model) tickCheckTimer() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return checkTimerMsg(t)
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Add debug logging
		//logger.Info("Key pressed in mode %d: %s", m.mode, msg.String())

		switch m.mode {
		case ModeNormal:
			// Only process mode-switching keys in normal mode
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "a":
				m.mode = ModeAddAccount
				m.textInput.Focus()
				return m, textinput.Blink
			case "l":
				m.mode = ModeListAccounts
			case "r":
				m.mode = ModeRemoveAccount
				m.textInput.Focus()
				m.textInput.Reset()
				return m, textinput.Blink
			}

		case ModeAddAccount:
			// In add mode, only handle enter and escape
			switch msg.String() {
			case "enter":
				return m, m.handleAddAccount(m.textInput.Value())
			case "esc":
				m.mode = ModeNormal
				m.error = nil
				m.textInput.Blur()
			}

		case ModeRemoveAccount:
			// In remove mode, handle navigation and selection
			switch msg.String() {
			case "enter":
				return m, m.handleRemoveByUsername(m.textInput.Value())
			case "esc":
				m.mode = ModeNormal
				m.error = nil
				m.textInput.Blur()
			}

		case ModeListAccounts:
			// In list mode, only handle escape
			if msg.String() == "esc" {
				m.mode = ModeNormal
				m.error = nil
			}
		}

	case checkTimerMsg:
		now := time.Now()
		elapsed := now.Sub(m.lastCheckTime)
		if elapsed >= m.checkInterval {
			logger.Info("Starting periodic check (interval: %s)", m.checkInterval)
			cmds = append(cmds, m.CheckAccounts())
			m.lastCheckTime = now
		}
		cmds = append(cmds, m.tickCheckTimer())

	case tickMsg:
		m.uptime = time.Since(m.startTime)
		cmds = append(cmds, m.tickUptime())

	case error:
		m.error = msg
		return m, nil

	default:
		// Handle spinner updates
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Handle text input updates only in add mode
	if m.mode == ModeAddAccount || m.mode == ModeRemoveAccount {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var s strings.Builder

	//s.WriteString(titleStyle.Render("X Track") + "\n\n")
	// Add status bar with spinner at the top
	s.WriteString(m.renderStatusBar() + "\n\n")


	// Main content area
	switch m.mode {
	case ModeAddAccount:
		prompt := inputPromptStyle.Render("Enter username to watch:")
		s.WriteString(prompt + " " + m.textInput.View() + "\n")
		s.WriteString(helpStyle.Render("\nPress enter to add, esc to cancel"))
	case ModeRemoveAccount:
		prompt := removePromptStyle.Render("Enter username to remove:")
		s.WriteString(prompt + " " + m.textInput.View() + "\n")
		s.WriteString(helpStyle.Render("\nPress enter to remove, esc to cancel"))
		s.WriteString(m.renderAccountList())
	case ModeListAccounts:
		s.WriteString(m.renderAccountList())
	}

	// Error display
	if m.error != nil {
		s.WriteString("\n" + errorStyle.Render(m.error.Error()))
	}

	// Help text
	s.WriteString("\n\n" + helpStyle.Render("a: add • l: list • r: remove • q: quit • esc: cancel"))

	return s.String()
}

func (m *Model) getModeString() string {
	switch m.mode {
	case ModeNormal:
		return "Normal"
	case ModeAddAccount:
		return "Add Account"
	case ModeListAccounts:
		return "List Accounts"
	case ModeRemoveAccount:
		return "Remove Account"
	default:
		return "Unknown"
	}
}

func (m *Model) renderAccountList() string {
	if len(m.accounts) == 0 {
		return "No accounts being watched"
	}

	var s strings.Builder
	s.WriteString("Watched accounts:\n\n")
	
	for _, account := range m.accounts {
		item := fmt.Sprintf("@%s",
			account.Username)
		s.WriteString(itemStyle.Render(item) + "\n")
	}
	
	return listStyle.Render(s.String())
}

func (m *Model) handleAddAccount(username string) tea.Cmd {
	return func() tea.Msg {
		// Remove @ if user added it anyway
		username = strings.TrimPrefix(username, "@")
		
		// Get user details from API
		user, err := m.api.GetUser(username)
		if err != nil {
			return err
		}

		logger.Info("Got user details - ID: %s, Username: %s, Following: %d", 
			user.RestID, 
			user.Legacy.ScreenName, 
			user.Legacy.FriendsCount)

		// Add to database
		account := &db.WatchedAccount{
			Username:        user.Legacy.ScreenName,
			UserID:         user.RestID,
		}

		if err := m.db.AddWatchedAccount(account); err != nil {
			return err
		}

		// Get and store initial following list
		followings, err := m.api.GetFollowingIDs(account.UserID)
		if err != nil {
			return fmt.Errorf("getting initial followings: %w", err)
		}

		if err := m.db.StoreFollowings(account.ID, followings.IDs); err != nil {
			return fmt.Errorf("storing initial followings: %w", err)
		}

		logger.Info("Initialized %d followings for @%s", len(followings.IDs), account.Username)

		m.mode = ModeNormal
		m.textInput.Reset()
		return m.loadAccounts()
	}
}

func (m *Model) handleRemoveByUsername(username string) tea.Cmd {
	return func() tea.Msg {
		// Remove @ if user added it anyway
		username = strings.TrimPrefix(username, "@")
		if username == "" {
			return fmt.Errorf("please enter a username")
		}
		
		// Find the account ID by username
		for _, account := range m.accounts {
			if account.Username == username {
				logger.Info("Removing account @%s (ID: %d)", username, account.ID)
				if err := m.db.RemoveWatchedAccount(account.ID); err != nil {
					return err
				}
				m.mode = ModeNormal
				m.textInput.Reset()
				m.textInput.Blur()
				return m.loadAccounts()
			}
		}
		return fmt.Errorf("account @%s not found", username)
	}
}

func (m *Model) loadAccounts() tea.Msg {
	accounts, err := m.db.GetWatchedAccounts()
	if err != nil {
		return err
	}
	m.accounts = accounts
	return nil
}

// CheckAccounts periodically checks all watched accounts for changes
func (m *Model) CheckAccounts() tea.Cmd {
	return tea.Tick(m.config.CheckInterval, func(t time.Time) tea.Msg {
		logger.Info("Starting periodic check of watched accounts...")
		
		accounts, err := m.db.GetWatchedAccounts()
		if err != nil {
			logger.Info("Error getting watched accounts: %v", err)
			return nil
		}

		for _, account := range accounts {
			// Get current following IDs from API
			followings, err := m.api.GetFollowingIDs(account.UserID)
			if err != nil {
				logger.Info("Error getting following IDs for %s: %v", account.Username, err)
				continue
			}

			// Get current followings from database
			currentFollowings, err := m.db.GetCurrentFollowings(account.ID)
			if err != nil {
				logger.Info("Error getting current followings for %s: %v", account.Username, err)
				continue
			}

			// Create map of new followings for efficient lookup
			newFollowingsMap := make(map[string]bool)
			var newFollows []string

			// Find new follows
			for _, id := range followings.IDs {
				newFollowingsMap[id] = true
				if !currentFollowings[id] {
					newFollows = append(newFollows, id)
				}
			}

			// Find unfollows
			var unfollows []string
			for id := range currentFollowings {
				if !newFollowingsMap[id] {
					unfollows = append(unfollows, id)
				}
			}

			// If there are changes, store them
			if len(newFollows) > 0 || len(unfollows) > 0 {
				logger.Info("Processing changes for %s: +%d new follows, -%d unfollows", 
					account.Username, len(newFollows), len(unfollows))

				// First store the events
				if err := m.db.StoreFollowEvents(account.ID, newFollows, unfollows); err != nil {
					logger.Info("Error storing follow events for %s: %v", account.Username, err)
					continue
				}

				// Then update the following relationships
				if err := m.db.StoreFollowings(account.ID, followings.IDs); err != nil {
					logger.Info("Error updating followings for %s: %v", account.Username, err)
					continue
				}

				// Send webhook notifications if configured
				if m.notifications != nil {
					// Handle follow notifications
					if m.config.EnableFollowNotifications && len(newFollows) > 0 {
						logger.Info("Sending follow notifications for %s: %d new follows", 
							account.Username, len(newFollows))
						m.notifications.NotifyNewFollows(&account, newFollows, m.api)
					} else if len(newFollows) > 0 {
						logger.Info("Follow notifications disabled, skipping %d new follows", len(newFollows))
					}

					// Handle unfollow notifications
					if m.config.EnableUnfollowNotifications && len(unfollows) > 0 {
						logger.Info("Sending unfollow notifications for %s: %d unfollows", 
							account.Username, len(unfollows))
						m.notifications.NotifyUnfollows(&account, unfollows, m.api)
					} else if len(unfollows) > 0 {
						logger.Info("Unfollow notifications disabled, skipping %d unfollows", len(unfollows))
					}
				}

				logger.Info("Successfully processed all changes for account %s", account.Username)
			} else {
				logger.Info("No changes detected for %s", account.Username)
			}
		}

		return CheckAccountsMsg(t)
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Helper function to format duration nicely
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// Add a helper function to print the current state
func (m *Model) debugState() {
	logger.Info("Current state - Mode: %d, Selected: %d, Accounts: %d", 
		m.mode, m.selected, len(m.accounts))
}

func (m *Model) renderStatusBar() string {
	uptime := time.Since(m.startTime).Round(time.Second)
	spinnerView := m.spinner.View()
	
	return statusBarStyle.Render(
		fmt.Sprintf("X Track | API Left: %d | Uptime: %s %s", 
			m.api.RemainingRequests(), 
			uptime, 
			spinnerView,
		),
	)
} 