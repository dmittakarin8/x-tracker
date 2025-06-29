# X-Tracker

A real-time X (Twitter) following tracker with interactive terminal interface and notification support.

## 🎯 Non-Technical Summary

**X-Tracker** is a simple command-line tool that helps you monitor who your favorite X (Twitter) accounts are following and unfollowing in real-time. Send alerts to discord / telegram. 

### What it does:
- **Monitors Accounts**: Watch any X account and track their following/unfollowing activity
- **Real-time Updates**: Automatically checks for changes at regular intervals (every 5 minutes by default)
- **Smart Notifications**: Get instant alerts via Discord or Telegram when someone follows or unfollows
- **Simple Interface**: Clean, colorful terminal interface that's easy to navigate
- **Data Storage**: Keeps a history of all changes in a local database

## 🚀 Features

- **Interactive TUI**: Terminal user interface built with Bubble Tea
- **Real-time Monitoring**: Automatic checking of following changes at configurable intervals
- **Multi-Platform Notifications**: Discord webhook and Telegram bot support
- **Local Database**: SQLite storage for persistent data and event history
- **Rate Limiting**: Smart API usage to respect X's rate limits
- **Configurable**: Environment-based configuration for easy customization
- **Logging**: Comprehensive logging system for debugging and monitoring
- **Cost Effective**:  Run this on a $3/month VPS and a $20/month RapidAPI account

## 📋 Prerequisites

- Go 1.21 or higher
- RapidAPI key for X (Twitter) API access
- (Optional) Discord webhook URL or Telegram bot token for notifications

## 🛠️ Installation

### From Source

1. Clone the repository:
```bash
git clone <repository-url>
cd x-tracker
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o x-tracker
```

4. Make it executable:
```bash
chmod +x x-tracker
```

## ⚙️ Configuration

Create a `.env` file in the project root with the following variables:

```env
# Required: RapidAPI Configuration
RAPID_API_KEY=your_rapidapi_key_here
RAPID_API_HOST=twitter154.p.rapidapi.com

# Optional: Notification Settings
DISCORD_WEBHOOK_URL=your_discord_webhook_url
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
TELEGRAM_CHAT_ID=your_telegram_chat_id

# Optional: Application Settings
CHECK_INTERVAL=5m
MAX_REQUESTS_PER_MINUTE=30
REQUEST_TIMEOUT=10s
LOGGING_ENABLED=true
LOG_DIR=~/.x-tracker/logs
DB_PATH=~/.x-tracker/data.db

# Optional: Notification Controls
ENABLE_FOLLOW_NOTIFICATIONS=true
ENABLE_UNFOLLOW_NOTIFICATIONS=true
ENABLE_DISCORD_NOTIFICATIONS=true
ENABLE_TELEGRAM_NOTIFICATIONS=true
```

### Getting API Keys

1. **RapidAPI Key**: 
   - Visit [RapidAPI](https://rapidapi.com)
   - Subscribe to the X (Twitter) API
   - Copy your API key

2. **Discord Webhook** (Optional):
   - Go to your Discord server settings
   - Navigate to Integrations → Webhooks
   - Create a new webhook and copy the URL

3. **Telegram Bot** (Optional):
   - Message [@BotFather](https://t.me/botfather) on Telegram
   - Create a new bot and get the token
   - Get your chat ID by messaging [@userinfobot](https://t.me/userinfobot)

## 🎮 Usage

### Starting the Application

```bash
./x-tracker
```

### Interactive Commands

Once the application is running, you can use these keyboard shortcuts:

- **`a`** - Add a new account to monitor
- **`l`** - List all monitored accounts
- **`r`** - Remove an account from monitoring
- **`q`** or **`Ctrl+C`** - Quit the application
- **`Esc`** - Return to main menu (when in sub-menus)

### Adding an Account

1. Press `a` to enter add mode
2. Type the username (without @) and press Enter
3. The account will be added to your monitoring list

### Viewing Accounts

Press `l` to see all accounts you're currently monitoring, along with their current following counts and last check times.

### Removing an Account

1. Press `r` to enter remove mode
2. Type the username and press Enter
3. The account will be removed from monitoring

## 🏗️ Architecture

The application follows a clean, modular architecture:

```
x-tracker/
├── main.go              # Application entry point
├── config/              # Configuration management
│   └── config.go
├── internal/            # Core application logic
│   ├── api/            # X API client
│   ├── db/             # Database operations
│   ├── ui/             # Terminal user interface
│   ├── webhook/        # Notification system
│   └── logger/         # Logging utilities
├── cmd/                # Command-line interface
└── go.mod              # Go module dependencies
```

### Key Components

- **API Client** (`internal/api/`): Handles all X API interactions with rate limiting
- **Database** (`internal/db/`): SQLite-based storage for accounts and events
- **UI** (`internal/ui/`): Bubble Tea-based terminal interface
- **Notifications** (`internal/webhook/`): Discord and Telegram integration
- **Configuration** (`config/`): Environment-based configuration management

## 🔧 Development

### Project Structure

The codebase is organized into logical packages:

- **`api`**: X API client with rate limiting and error handling
- **`db`**: Database models and operations for accounts and events
- **`ui`**: Terminal UI components and state management
- **`webhook`**: Notification system for Discord and Telegram
- **`logger`**: Structured logging with file and console output
- **`config`**: Configuration loading and validation

### Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o x-tracker

# macOS
GOOS=darwin GOARCH=amd64 go build -o x-tracker

# Windows
GOOS=windows GOARCH=amd64 go build -o x-tracker.exe
```

### Running Tests

```bash
go test ./...
```

## 📊 Data Storage

The application uses SQLite for data persistence:

- **Watched Accounts**: List of accounts being monitored
- **Followed Accounts**: Current following relationships
- **Follow Events**: Historical record of follow/unfollow events

Database location: `~/.x-tracker/data.db` (configurable)

## 🔔 Notifications

### Discord Notifications

When enabled, the application sends rich Discord embeds containing:
- Account information (username, profile picture)
- List of new follows/unfollows with usernames
- Timestamps and event details

### Telegram Notifications

Telegram notifications include:
- Account username and profile information
- Formatted lists of follows/unfollows
- Direct links to X profiles

## 🐛 Troubleshooting

### Common Issues

1. **API Rate Limiting**: 
   - Reduce `MAX_REQUESTS_PER_MINUTE` in your `.env`
   - Increase `CHECK_INTERVAL` to check less frequently

2. **Database Errors**:
   - Check file permissions for `~/.x-tracker/`
   - Ensure sufficient disk space

3. **Notification Failures**:
   - Verify webhook URLs and bot tokens
   - Check network connectivity
   - Review logs for specific error messages

### Logs

Enable logging by setting `LOGGING_ENABLED=true` in your `.env` file. Logs are stored in `~/.x-tracker/logs/` by default.

## 📝 License

[Add your license information here]

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📞 Support

For issues and questions:
- Create an issue on GitHub
- Check the troubleshooting section
- Review the logs for error details

---

**Note**: This tool is for educational and personal use. Please respect X's Terms of Service and API rate limits when using this application. 