package auth

import (
	"time"

	"github.com/google/uuid"
)

// User is the domain model for the users table.
type User struct {
	ID           uuid.UUID
	Email        string
	FullName     string
	PasswordHash *string
	CreatedAt    time.Time
}

// SessionMeta is the request metadata recorded with a new session.
type SessionMeta struct {
	IPAddress *string
	UserAgent *string
}

// LoginInput is what the service needs to log a user in with a password.
type LoginInput struct {
	Email    string
	Password string
	Meta     SessionMeta
}

// RegisterInput is what the service needs to register a user with a password.
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

// Result is the user and the raw token of the session just created for them.
type Result struct {
	User  User
	Token string
}

// newSession is what the repository needs to insert a session.
type newSession struct {
	TokenHash string
	ExpiresAt time.Time
	Meta      SessionMeta
}
