package db

import (
	"time"
)

type WatchedAccount struct {
	ID       int64  `db:"id"`
	Username string `db:"username"`
	UserID   string `db:"user_id"`
}

type FollowedAccount struct {
	WatchedAccountID int64  `db:"watched_account_id"`
	UserID          string `db:"followed_user_id"`
}

type EventType string

const (
	EventTypeFollow   EventType = "follow"
	EventTypeUnfollow EventType = "unfollow"
)

type FollowEvent struct {
	ID              int64     `db:"id"`
	WatchedAccountID int64     `db:"watched_account_id"`
	UserID          string    `db:"user_id"`
	EventType       EventType `db:"event_type"`
	DetectedAt      time.Time `db:"detected_at"`
} 