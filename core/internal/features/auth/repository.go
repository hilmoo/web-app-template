package auth

import (
	"context"
	"errors"
	"fmt"

	db "web-app-template-core/internal/db/sqlc"
	"web-app-template-core/internal/helper"

	"github.com/google/uuid"
	"github.com/hilmoo/gomox/sqlx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data-access contract for users and their sessions.
type Repository interface {
	FindUserByID(ctx context.Context, id uuid.UUID) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, in UpdateMeInput) (User, error)
	CreateUserWithSession(ctx context.Context, u newUser, s newSession) (User, error)
	CreateSession(ctx context.Context, userID uuid.UUID, s newSession) error
	FindSessionUserID(ctx context.Context, tokenHash string) (uuid.UUID, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error
}

type postgresRepo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepo{pool: pool, q: db.New(pool)}
}

func (r *postgresRepo) FindUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row, err := r.q.GetUser(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, userNotFound(id.String())
	}
	if err != nil {
		return User{}, fmt.Errorf("auth: find user: %w", err)
	}
	return userFromRow(row), nil
}

func (r *postgresRepo) FindUserByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, userNotFound(email)
	}
	if err != nil {
		return User{}, fmt.Errorf("auth: find user by email: %w", err)
	}
	return userFromRow(row), nil
}

func (r *postgresRepo) UpdateUser(ctx context.Context, id uuid.UUID, in UpdateMeInput) (User, error) {
	row, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:           id,
		Email:        in.Email,
		FullName:     in.FullName,
		PasswordHash: nil,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, userNotFound(id.String())
	}
	if err != nil {
		if errors.Is(sqlx.ClassifyPgError(err), sqlx.ErrAlreadyExists) {
			return User{}, fmt.Errorf("%w: %w", ErrEmailAlreadyRegistered, err)
		}
		return User{}, fmt.Errorf("auth: update user: %w", err)
	}
	return userFromRow(row), nil
}

// CreateUserWithSession creates the user and its first session atomically, so a failed
// session insert cannot leave a registered user behind without a session.
func (r *postgresRepo) CreateUserWithSession(ctx context.Context, u newUser, s newSession) (User, error) {
	return helper.WithTx(ctx, r.pool, r.q, func(q *db.Queries) (User, error) {
		row, err := q.CreateUser(ctx, db.CreateUserParams{
			Email:        u.Email,
			FullName:     u.FullName,
			PasswordHash: &u.PasswordHash,
		})
		if err != nil {
			if errors.Is(sqlx.ClassifyPgError(err), sqlx.ErrAlreadyExists) {
				return User{}, fmt.Errorf("%w: %w", ErrEmailAlreadyRegistered, err)
			}
			return User{}, fmt.Errorf("auth: create user: %w", err)
		}
		if err = createSession(ctx, q, row.ID, s); err != nil {
			return User{}, err
		}
		return userFromRow(row), nil
	})
}

func (r *postgresRepo) CreateSession(ctx context.Context, userID uuid.UUID, s newSession) error {
	return createSession(ctx, r.q, userID, s)
}

func (r *postgresRepo) FindSessionUserID(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	row, err := r.q.GetSessionByTokenHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrSessionNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth: find session: %w", err)
	}
	return row.UserID, nil
}

func (r *postgresRepo) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	if err := r.q.DeleteSessionByTokenHash(ctx, tokenHash); err != nil {
		return fmt.Errorf("auth: delete session: %w", err)
	}
	return nil
}

func createSession(ctx context.Context, q *db.Queries, userID uuid.UUID, s newSession) error {
	err := q.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		TokenHash: s.TokenHash,
		ExpiresAt: s.ExpiresAt,
		IpAddress: s.Meta.IPAddress,
		UserAgent: s.Meta.UserAgent,
	})
	if err != nil {
		return fmt.Errorf("auth: create session: %w", err)
	}
	return nil
}

func userFromRow(r db.User) User {
	return User{
		ID:           r.ID,
		Email:        r.Email,
		FullName:     r.FullName,
		PasswordHash: r.PasswordHash,
		CreatedAt:    r.CreatedAt,
	}
}
