package auth //nolint:testpackage // exercises unexported service internals

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory Repository keyed by email and session token hash.
type fakeRepo struct {
	users    map[string]User
	sessions map[string]uuid.UUID
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: map[string]User{}, sessions: map[string]uuid.UUID{}}
}

func (f *fakeRepo) FindUserByID(_ context.Context, id uuid.UUID) (User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return User{}, ErrUserNotFound
}

func (f *fakeRepo) FindUserByEmail(_ context.Context, email string) (User, error) {
	u, ok := f.users[email]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}

func (f *fakeRepo) UpdateUser(ctx context.Context, id uuid.UUID, in UpdateMeInput) (User, error) {
	u, err := f.FindUserByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	if in.FullName != nil {
		u.FullName = *in.FullName
	}
	f.users[u.Email] = u
	return u, nil
}

func (f *fakeRepo) CreateUserWithSession(
	ctx context.Context,
	email, fullName, passwordHash string,
	s newSession,
) (User, error) {
	if _, ok := f.users[email]; ok {
		return User{}, ErrEmailAlreadyRegistered
	}
	user := User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     fullName,
		PasswordHash: &passwordHash,
		CreatedAt:    time.Now(),
	}
	f.users[email] = user
	return user, f.CreateSession(ctx, user.ID, s)
}

func (f *fakeRepo) CreateSession(_ context.Context, userID uuid.UUID, s newSession) error {
	f.sessions[s.TokenHash] = userID
	return nil
}

func (f *fakeRepo) FindSessionUserID(_ context.Context, tokenHash string) (uuid.UUID, error) {
	id, ok := f.sessions[tokenHash]
	if !ok {
		return uuid.Nil, ErrSessionNotFound
	}
	return id, nil
}

func (f *fakeRepo) DeleteSessionByTokenHash(_ context.Context, tokenHash string) error {
	delete(f.sessions, tokenHash)
	return nil
}

func TestServiceRegisterLoginLogout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewService(newFakeRepo(), "secret")

	reg, err := svc.Register(ctx, registerIn("a@example.com", "pw"))
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if got, lookupErr := svc.UserIDByToken(ctx, reg.Token); lookupErr != nil || got != reg.User.ID {
		t.Fatalf("UserIDByToken(register token) = %v, %v; want %v", got, lookupErr, reg.User.ID)
	}

	_, err = svc.Register(ctx, registerIn("a@example.com", "pw"))
	if !errors.Is(err, ErrEmailAlreadyRegistered) {
		t.Errorf("duplicate Register() error = %v, want ErrEmailAlreadyRegistered", err)
	}

	login, err := svc.Login(ctx, loginIn("a@example.com", "pw"))
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if login.User.ID != reg.User.ID {
		t.Errorf("Login() user = %v, want %v", login.User.ID, reg.User.ID)
	}

	if err = svc.Logout(ctx, login.Token); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, err = svc.UserIDByToken(ctx, login.Token); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("UserIDByToken(after logout) error = %v, want ErrSessionNotFound", err)
	}
}

func TestServiceLoginInvalidCredentials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, "secret")

	if _, err := svc.Register(ctx, registerIn("a@example.com", "pw")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	repo.users["nopw@example.com"] = User{
		ID:           uuid.New(),
		Email:        "nopw@example.com",
		FullName:     "",
		PasswordHash: nil,
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name string
		in   LoginInput
	}{
		{"unknown email", loginIn("missing@example.com", "pw")},
		{"wrong password", loginIn("a@example.com", "nope")},
		{"no password set", loginIn("nopw@example.com", "pw")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := svc.Login(ctx, tt.in); !errors.Is(err, ErrInvalidCredentials) {
				t.Errorf("Login() error = %v, want ErrInvalidCredentials", err)
			}
		})
	}
}

func TestToAppErrorHidesInternals(t *testing.T) {
	t.Parallel()

	errH := toAppError(errors.New("pg: connection refused"))
	if errH.ReasonField != "internal error" {
		t.Errorf("reason = %q, want generic message", errH.ReasonField)
	}
	if got := toAppError(ErrUserNotFound).CodeField; got != 404 {
		t.Errorf("user not found code = %d, want 404", got)
	}
}

func registerIn(email, password string) RegisterInput {
	return RegisterInput{
		Email:    email,
		FullName: "A",
		Password: password,
		Meta:     SessionMeta{IPAddress: nil, UserAgent: nil},
	}
}

func loginIn(email, password string) LoginInput {
	return LoginInput{Email: email, Password: password, Meta: SessionMeta{IPAddress: nil, UserAgent: nil}}
}
