package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	_ "github.com/mattn/go-sqlite3"
	"x-tracker/internal/logger"
	"time"
)

type Database struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS watched_accounts (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE,
    user_id TEXT
);

CREATE TABLE IF NOT EXISTS following (
    watched_account_id INTEGER,
    followed_user_id TEXT,
    PRIMARY KEY (watched_account_id, followed_user_id),
    FOREIGN KEY(watched_account_id) REFERENCES watched_accounts(id)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS follow_events (
    id INTEGER PRIMARY KEY,
    watched_account_id INTEGER,
    user_id TEXT,
    event_type TEXT CHECK(event_type IN ('follow', 'unfollow')),
    detected_at TIMESTAMP,
    FOREIGN KEY(watched_account_id) REFERENCES watched_accounts(id)
);

CREATE INDEX IF NOT EXISTS idx_follow_events_account 
ON follow_events(watched_account_id, detected_at);`

func NewDatabase(dbPath string) (*Database, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	// Initialize schema
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// AddWatchedAccount adds a new account to watch
func (d *Database) AddWatchedAccount(account *WatchedAccount) error {
	logger.Info("Adding account to watch list: %s", account.Username)
	query := `
		INSERT INTO watched_accounts (username, user_id)
		VALUES (?, ?)`
	
	result, err := d.db.Exec(query,
		account.Username,
		account.UserID)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	account.ID = id
	
	logger.Info("Successfully added account: %s (ID: %d)", account.Username, account.ID)
	return nil
}

// GetWatchedAccounts returns all watched accounts
func (d *Database) GetWatchedAccounts() ([]WatchedAccount, error) {
	var accounts []WatchedAccount
	rows, err := d.db.Query(`
		SELECT id, username, user_id 
		FROM watched_accounts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var account WatchedAccount
		err := rows.Scan(
			&account.ID,
			&account.Username,
			&account.UserID)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// RemoveWatchedAccount removes a watched account
func (d *Database) RemoveWatchedAccount(id int64) error {
	logger.Info("Removing watched account ID: %d", id)
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete from following table first (foreign key constraint)
	if _, err := tx.Exec("DELETE FROM following WHERE watched_account_id = ?", id); err != nil {
		return err
	}

	// Delete from watched_accounts
	if _, err := tx.Exec("DELETE FROM watched_accounts WHERE id = ?", id); err != nil {
		return err
	}

	return tx.Commit()
	
	logger.Info("Successfully removed account ID: %d", id)
	return nil
}

// StoreFollowings stores multiple following relationships
func (d *Database) StoreFollowings(watchedAccountID int64, followingIDs []string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current followings
	currentFollowings, err := d.GetCurrentFollowings(watchedAccountID)
	if err != nil {
		return fmt.Errorf("getting current followings: %w", err)
	}

	// Create map of new followings for efficient lookup
	newFollowingsMap := make(map[string]bool)
	for _, id := range followingIDs {
		newFollowingsMap[id] = true
	}

	// Find and delete unfollows
	for id := range currentFollowings {
		if !newFollowingsMap[id] {
			_, err = tx.Exec("DELETE FROM following WHERE watched_account_id = ? AND followed_user_id = ?", 
				watchedAccountID, id)
			if err != nil {
				return fmt.Errorf("deleting unfollow %s: %w", id, err)
			}
			logger.Info("Removed following relationship: account %d -> user %s", watchedAccountID, id)
		}
	}

	// Insert only new follows
	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO following 
		(watched_account_id, followed_user_id)
		VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	// Insert each new following relationship
	for _, id := range followingIDs {
		if !currentFollowings[id] {
			_, err := stmt.Exec(watchedAccountID, id)
			if err != nil {
				return fmt.Errorf("inserting new follow %s: %w", id, err)
			}
			//logger.Info("Added new following relationship: account %d -> user %s", watchedAccountID, id)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	logger.Info("Updated following relationships for account ID %d", watchedAccountID)
	return nil
}

// GetCurrentFollowings gets all current following IDs for an account
func (d *Database) GetCurrentFollowings(watchedAccountID int64) (map[string]bool, error) {
	rows, err := d.db.Query(
		"SELECT followed_user_id FROM following WHERE watched_account_id = ?",
		watchedAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	followings := make(map[string]bool)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		followings[userID] = true
	}
	return followings, nil
}

// StoreFollowEvents records follow/unfollow events
func (d *Database) StoreFollowEvents(watchedAccountID int64, follows, unfollows []string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO follow_events 
		(watched_account_id, user_id, event_type, detected_at)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()

	// Store new follows
	for _, userID := range follows {
		_, err := stmt.Exec(watchedAccountID, userID, EventTypeFollow, now)
		if err != nil {
			return fmt.Errorf("inserting follow event for %s: %w", userID, err)
		}
		logger.Info("Stored follow event for account %d: following %s", watchedAccountID, userID)
	}

	// Store unfollows
	for _, userID := range unfollows {
		_, err := stmt.Exec(watchedAccountID, userID, EventTypeUnfollow, now)
		if err != nil {
			return fmt.Errorf("inserting unfollow event for %s: %w", userID, err)
		}
		logger.Info("Stored unfollow event for account %d: unfollowed %s", watchedAccountID, userID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	logger.Info("Successfully stored %d follow and %d unfollow events", len(follows), len(unfollows))
	return nil
}

// ProcessFollowingChanges detects and stores following changes
func (d *Database) ProcessFollowingChanges(account *WatchedAccount, newFollowingIDs []string) error {
	// Get current followings
	currentFollowings, err := d.GetCurrentFollowings(account.ID)
	if err != nil {
		return fmt.Errorf("getting current followings: %w", err)
	}

	logger.Info("Current followings in DB for %s: %d, New followings from API: %d", 
		account.Username, len(currentFollowings), len(newFollowingIDs))

	// Track changes
	var newFollows []string
	newFollowingsMap := make(map[string]bool)

	// Debug: Log all current following IDs
	//logger.Info("Current following IDs in DB for %s: %v", account.Username, currentFollowings)
	
	// Debug: Log all new following IDs
	//logger.Info("New following IDs from API for %s: %v", account.Username, newFollowingIDs)

	// Find new follows
	for _, id := range newFollowingIDs {
		newFollowingsMap[id] = true
		if !currentFollowings[id] {
			logger.Info("Found new follow: %s", id)
			newFollows = append(newFollows, id)
		}
	}

	// Find unfollows
	var unfollows []string
	for id := range currentFollowings {
		if !newFollowingsMap[id] {
			logger.Info("Found unfollow: %s", id)
			unfollows = append(unfollows, id)
		}
	}

	// If there are changes, store them
	if len(newFollows) > 0 || len(unfollows) > 0 {
		logger.Info("Processing changes for %s: +%d new follows, -%d unfollows", 
			account.Username, len(newFollows), len(unfollows))

		// First store the events
		if err := d.StoreFollowEvents(account.ID, newFollows, unfollows); err != nil {
			return fmt.Errorf("storing follow events: %w", err)
		}

		// Then update the following relationships
		if err := d.StoreFollowings(account.ID, newFollowingIDs); err != nil {
			return fmt.Errorf("updating followings: %w", err)
		}

		logger.Info("Successfully processed all changes for account %s", account.Username)
	} else {
		logger.Info("No changes detected for %s", account.Username)
	}

	return nil
} 