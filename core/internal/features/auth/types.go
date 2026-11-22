package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	FullName     string
	PasswordHash *string
	CreatedAt    time.Time
}

type SessionMeta struct {
	IPAddress *string
	UserAgent *string
}

type LoginInput struct {
	Email    string
	Password string
	Meta     SessionMeta
}

type RegisterInput struct {
	Email    string
	FullName string
	Password string
	Meta     SessionMeta
}

// UpdateMeInput is a partial update of the caller; nil fields are left unchanged.
type UpdateMeInput struct {
	Email    *string
	FullName *string
}

type Result struct {
	User  User
	Token string
}

type newSession struct {
	TokenHash string
	ExpiresAt time.Time
	Meta      SessionMeta
}
