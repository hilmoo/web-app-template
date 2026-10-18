package auth

import (
	"context"

	db "web-app-template-core/internal/db/sqlc"
	"web-app-template-core/internal/helper"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data-access contract for users and their sessions.
type Repository interface {
	FindUserByID(ctx context.Context, id uuid.UUID) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, in UpdateMeInput) (User, error)
	CreateUserWithSession(ctx context.Context, email, fullName, passwordHash string, s newSession) (User, error)
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
	if err != nil {
		return User{}, helper.MapDBError(err, "auth: find user", ErrUserNotFound, nil)
	}
	return userFromRow(row), nil
}

func (r *postgresRepo) FindUserByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, helper.MapDBError(err, "auth: find user by email", ErrUserNotFound, nil)
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
	if err != nil {
		return User{}, helper.MapDBError(err, "auth: update user", ErrUserNotFound, ErrEmailAlreadyRegistered)
	}
	return userFromRow(row), nil
}

// CreateUserWithSession creates the user and its first session atomically, so a failed
// session insert cannot leave a registered user behind without a session.
func (r *postgresRepo) CreateUserWithSession(
	ctx context.Context,
	email, fullName, passwordHash string,
	s newSession,
) (User, error) {
	return helper.WithTx(ctx, r.pool, r.q, func(q *db.Queries) (User, error) {
		row, err := q.CreateUser(ctx, db.CreateUserParams{
			Email:        email,
			FullName:     fullName,
			PasswordHash: &passwordHash,
		})
		if err != nil {
			return User{}, helper.MapDBError(err, "auth: create user", nil, ErrEmailAlreadyRegistered)
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
	if err != nil {
		return uuid.Nil, helper.MapDBError(err, "auth: find session", ErrSessionNotFound, nil)
	}
	return row.UserID, nil
}

func (r *postgresRepo) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	return helper.MapDBError(r.q.DeleteSessionByTokenHash(ctx, tokenHash), "auth: delete session", nil, nil)
}

func createSession(ctx context.Context, q *db.Queries, userID uuid.UUID, s newSession) error {
	err := q.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		TokenHash: s.TokenHash,
		ExpiresAt: s.ExpiresAt,
		IpAddress: s.Meta.IPAddress,
		UserAgent: s.Meta.UserAgent,
	})
	return helper.MapDBError(err, "auth: create session", nil, nil)
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
