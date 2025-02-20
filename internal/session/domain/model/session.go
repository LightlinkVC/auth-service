package model

import "time"

type Session struct {
	JWTAccess        string    `json:"access_token"`
	JWTRefresh       string    `json:"refresh_token"`
	UserID           uint      `json:"user_id"`
	Username         string    `json:"username"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}
