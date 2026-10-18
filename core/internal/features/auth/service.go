package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/matthewhartstonge/argon2"
)

// Service defines the auth business operations.
type Service interface {
	Login(ctx context.Context, in LoginInput) (Result, error)
	Register(ctx context.Context, in RegisterInput) (Result, error)
	Logout(ctx context.Context, token string) error
	Me(ctx context.Context, userID uuid.UUID) (User, error)
	UpdateMe(ctx context.Context, userID uuid.UUID, in UpdateMeInput) (User, error)
	// UserIDByToken satisfies helper.SessionResolver for the auth interceptor.
	UserIDByToken(ctx context.Context, token string) (uuid.UUID, error)
}

type service struct {
	repo   Repository
	secret string
	now    func() time.Time
}

func NewService(repo Repository, secret string) Service {
	return &service{repo: repo, secret: secret, now: time.Now}
}

func (s *service) Login(ctx context.Context, in LoginInput) (Result, error) {
	user, err := s.repo.FindUserByEmail(ctx, in.Email)
	if errors.Is(err, ErrUserNotFound) {
		return Result{}, ErrInvalidCredentials // don't reveal which part was wrong
	}
	if err != nil {
		return Result{}, err
	}
	if user.PasswordHash == nil {
		return Result{}, ErrInvalidCredentials
	}

	ok, err := argon2.VerifyEncoded([]byte(in.Password), []byte(*user.PasswordHash))
	if err != nil {
		return Result{}, fmt.Errorf("auth: verify password: %w", err)
	}
	if !ok {
		return Result{}, ErrInvalidCredentials
	}

	token, sess, err := s.newSession(in.Meta)
	if err != nil {
		return Result{}, err
	}
	if err = s.repo.CreateSession(ctx, user.ID, sess); err != nil {
		return Result{}, err
	}

	return Result{User: user, Token: token}, nil
}

func (s *service) Register(ctx context.Context, in RegisterInput) (Result, error) {
	hash, err := hashPasswordArgon(in.Password)
	if err != nil {
		return Result{}, err
	}

	token, sess, err := s.newSession(in.Meta)
	if err != nil {
		return Result{}, err
	}

	user, err := s.repo.CreateUserWithSession(ctx, in.Email, in.FullName, hash, sess)
	if err != nil {
		return Result{}, err
	}

	return Result{User: user, Token: token}, nil
}

func (s *service) Logout(ctx context.Context, token string) error {
	return s.repo.DeleteSessionByTokenHash(ctx, hashSessionToken(s.secret, token))
}

func (s *service) Me(ctx context.Context, userID uuid.UUID) (User, error) {
	return s.repo.FindUserByID(ctx, userID)
}

func (s *service) UpdateMe(ctx context.Context, userID uuid.UUID, in UpdateMeInput) (User, error) {
	return s.repo.UpdateUser(ctx, userID, in)
}

func (s *service) UserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	return s.repo.FindSessionUserID(ctx, hashSessionToken(s.secret, token))
}

// newSession returns a fresh raw token and the session row that stores its hash.
func (s *service) newSession(meta SessionMeta) (string, newSession, error) {
	token, err := newSessionToken()
	if err != nil {
		return "", newSession{}, err
	}
	return token, newSession{
		TokenHash: hashSessionToken(s.secret, token),
		ExpiresAt: s.now().Add(sessionDuration),
		Meta:      meta,
	}, nil
}
