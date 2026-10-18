package helper_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	"web-app-template-core/internal/config"
	"web-app-template-core/internal/helper"
)

const (
	testProcedure     = "/test.v1/Do"
	publicProcedure   = "/test.v1/Public"
	internalProcedure = "/test.v1/Internal"
	serviceSecret     = "service-secret"
)

// fakeSessions is a SessionResolver backed by a token -> user map.
type fakeSessions map[string]uuid.UUID

func (f fakeSessions) UserIDByToken(_ context.Context, token string) (uuid.UUID, error) {
	id, ok := f[token]
	if !ok {
		return uuid.Nil, errors.New("no such session")
	}

	return id, nil
}

// newClient serves one procedure behind the interceptor; seen receives the caller's user id.
func newClient(
	t *testing.T,
	sessions helper.SessionResolver,
	procedure string,
	seen *uuid.UUID,
) *connect.Client[emptypb.Empty, emptypb.Empty] {
	t.Helper()

	interceptor := helper.NewAuthInterceptor(
		helper.HandlerParams{
			Logger:   nil,
			Validate: nil,
			Pool:     nil,
			Queries:  nil,
			Config:   config.Config{ServiceSecret: serviceSecret},
		},
		sessions,
		helper.AccessPolicy{
			publicProcedure:   helper.AccessPublic,
			internalProcedure: helper.AccessService,
		},
	)

	mux := http.NewServeMux()
	mux.Handle(procedure, connect.NewUnaryHandler(
		procedure,
		func(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
			if id, errH := helper.UserIDFromContext(ctx); errH == nil {
				*seen = id
			}

			return connect.NewResponse(&emptypb.Empty{}), nil
		},
		connect.WithInterceptors(interceptor),
	))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return connect.NewClient[emptypb.Empty, emptypb.Empty](srv.Client(), srv.URL+procedure)
}

func call(c *connect.Client[emptypb.Empty, emptypb.Empty], headers map[string]string) error {
	req := connect.NewRequest(&emptypb.Empty{})
	for k, v := range headers {
		req.Header().Set(k, v)
	}

	_, err := c.CallUnary(context.Background(), req)

	return err
}

func TestAuthInterceptor(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	sessions := fakeSessions{"good": userID}

	tests := []struct {
		name      string
		procedure string
		headers   map[string]string
		wantCode  connect.Code // 0 means success
		wantUser  uuid.UUID
	}{
		{"valid session", testProcedure, map[string]string{"Cookie": "session=good"}, 0, userID},
		{
			"unknown session",
			testProcedure,
			map[string]string{"Cookie": "session=bad"},
			connect.CodeUnauthenticated,
			uuid.Nil,
		},
		{"no credentials", testProcedure, nil, connect.CodeUnauthenticated, uuid.Nil},
		{"public without credentials", publicProcedure, nil, 0, uuid.Nil},
		{"public ignores stale cookie", publicProcedure, map[string]string{"Cookie": "session=bad"}, 0, uuid.Nil},
		{"service secret", testProcedure, map[string]string{"Authorization": "Bearer " + serviceSecret}, 0, uuid.Nil},
		{
			"wrong service secret",
			testProcedure,
			map[string]string{"Authorization": "Bearer nope"},
			connect.CodeUnauthenticated,
			uuid.Nil,
		},
		{"service acts as user", testProcedure, map[string]string{
			"Authorization":            "Bearer " + serviceSecret,
			helper.ServiceUserIDHeader: userID.String(),
		}, 0, userID},
		{
			"user session on service-only",
			internalProcedure,
			map[string]string{"Cookie": "session=good"},
			connect.CodePermissionDenied,
			uuid.Nil,
		},
		{
			"service on service-only",
			internalProcedure,
			map[string]string{"Authorization": "Bearer " + serviceSecret},
			0,
			uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var seen uuid.UUID

			err := call(newClient(t, sessions, tt.procedure, &seen), tt.headers)

			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if got := connect.CodeOf(err); got != tt.wantCode {
				t.Fatalf("code = %v (err %v), want %v", got, err, tt.wantCode)
			}

			if seen != tt.wantUser {
				t.Errorf("handler saw user %v, want %v", seen, tt.wantUser)
			}
		})
	}
}
