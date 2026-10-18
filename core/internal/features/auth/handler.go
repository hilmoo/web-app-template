package auth

import (
	"context"
	servicesv1 "proto/services/v1"
	"proto/services/v1/servicesv1connect"

	db "web-app-template-core/internal/gen/sqlc"
	"web-app-template-core/internal/helper"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	errort "github.com/hilmoo/gomox/transport/error"
	"github.com/hilmoo/gomox/transport/validation"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/herodot"
)

type connectHandler struct {
	servicesv1connect.UnimplementedAuthServiceHandler

	validate *validation.Vld
	pool     *pgxpool.Pool
	queries  *db.Queries
	secret   string
	isProd   bool
}

var _ servicesv1connect.AuthServiceHandler = (*connectHandler)(nil)

func NewConnect(arg helper.HandlerParams) servicesv1connect.AuthServiceHandler {
	return &connectHandler{
		validate: arg.Validate,
		pool:     arg.Pool,
		queries:  arg.Queries,
		secret:   arg.Config.Secret,
		isProd:   arg.Config.IsProd,
	}
}

func (h *connectHandler) AuthLogin(
	ctx context.Context,
	req *servicesv1.AuthLoginRequest,
) (*servicesv1.AuthLoginResponse, error) {
	callInfo, errH := handlerCallInfo(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	err := h.validate.ValidatePayload(req)
	if err != nil {
		return nil, errort.ConnectError(ctx, herodot.ErrBadRequest.WithReason("invalid request").WithDebug(err.Error()))
	}

	resp, errH := passwordLoginService(ctx, passwordLoginServiceParams{
		queries:   h.queries,
		secret:    h.secret,
		in:        req,
		ipAddress: peerIP(callInfo.Peer().Addr),
		userAgent: strPtr(callInfo.RequestHeader().Get("User-Agent")),
	})
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	setSessionCookie(callInfo.ResponseHeader(), resp.token, h.isProd)
	return &servicesv1.AuthLoginResponse{
		Data: toMeUser(resp.user),
	}, nil
}

func (h *connectHandler) AuthRegister(
	ctx context.Context,
	req *servicesv1.AuthRegisterRequest,
) (*servicesv1.AuthRegisterResponse, error) {
	callInfo, errH := handlerCallInfo(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	if errH = h.validate.ValidatePayload(req); errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	resp, errH := passwordRegisterService(ctx, passwordRegisterServiceParams{
		pool:      h.pool,
		queries:   h.queries,
		secret:    h.secret,
		in:        req,
		ipAddress: peerIP(callInfo.Peer().Addr),
		userAgent: strPtr(callInfo.RequestHeader().Get("User-Agent")),
	})
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	setSessionCookie(callInfo.ResponseHeader(), resp.token, h.isProd)
	return &servicesv1.AuthRegisterResponse{Data: toMeUser(resp.user)}, nil
}

func (h *connectHandler) AuthLogout(
	ctx context.Context,
	_ *servicesv1.AuthLogoutRequest,
) (*servicesv1.AuthLogoutResponse, error) {
	callInfo, errH := handlerCallInfo(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	token, errH := helper.SessionTokenFromHeader(callInfo.RequestHeader())
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	errH = passwordLogoutService(ctx, passwordLogoutServiceParams{
		queries: h.queries,
		secret:  h.secret,
		token:   token,
	})
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	clearSessionCookie(callInfo.ResponseHeader(), h.isProd)
	return &servicesv1.AuthLogoutResponse{}, nil
}

func (h *connectHandler) AuthMe(
	ctx context.Context,
	_ *servicesv1.AuthMeRequest,
) (*servicesv1.AuthMeResponse, error) {
	return helper.WithUserID(ctx, func(userID uuid.UUID) (*servicesv1.AuthMeResponse, *herodot.DefaultError) {
		user, errH := meGetUser(ctx, h.queries, userID)
		if errH != nil {
			return nil, errH
		}

		return &servicesv1.AuthMeResponse{Data: toMeUser(user)}, nil
	})
}

func (h *connectHandler) AuthUpdateMe(
	ctx context.Context,
	req *servicesv1.AuthUpdateMeRequest,
) (*servicesv1.AuthUpdateMeResponse, error) {
	return helper.WithUserID(ctx, func(userID uuid.UUID) (*servicesv1.AuthUpdateMeResponse, *herodot.DefaultError) {
		updated, errH := meUpdateUserService(ctx, meUpdateUserServiceParams{
			queries: h.queries,
			userID:  userID,
			in:      req,
		})
		if errH != nil {
			return nil, errH
		}

		return &servicesv1.AuthUpdateMeResponse{Data: toMeUser(updated)}, nil
	})
}

func handlerCallInfo(ctx context.Context) (connect.CallInfo, *herodot.DefaultError) {
	callInfo, ok := connect.CallInfoForHandlerContext(ctx)
	if !ok {
		return nil, herodot.ErrInternalServerError.WithReason("missing request metadata")
	}
	return callInfo, nil
}
