package api

// User by Name Response 
type UserResponse struct {
	RestID string `json:"rest_id"`
	Legacy struct {
		CreatedAt           string `json:"created_at"`
		Name               string `json:"name"`
		ScreenName         string `json:"screen_name"`
		FriendsCount       int    `json:"friends_count"`      // This is following_count
		FollowersCount     int    `json:"followers_count"`
		FavouritesCount    int    `json:"favourites_count"`
		ProfileImageURLHTTPS string `json:"profile_image_url_https"`
		Verified           bool   `json:"verified"`
	} `json:"legacy"`
	IsBlueVerified bool `json:"is_blue_verified"`
}

// FollowingIDsResponse represents the API response for following IDs
type FollowingIDsResponse struct {
	IDs                []string `json:"ids"`
	NextCursor         int64    `json:"next_cursor"`
	NextCursorStr      string   `json:"next_cursor_str"`
	PreviousCursor     int64    `json:"previous_cursor"`
	PreviousCursorStr  string   `json:"previous_cursor_str"`
	TotalCount         *int     `json:"total_count"`
}

// UserByIDResponse represents the API response for user lookup by ID
type UserByIDResponse struct {
	RestID string `json:"rest_id"`
	Legacy struct {
		ScreenName string `json:"screen_name"`
		Name       string `json:"name"`
		FollowersCount     int    `json:"followers_count"`
	} `json:"legacy"`
} 