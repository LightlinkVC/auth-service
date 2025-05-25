package entity

import "time"

type Session struct {
	JWTAccess        string
	JWTRefresh       string
	UserID           uint
	Username         string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}
