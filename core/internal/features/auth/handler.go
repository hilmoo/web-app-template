package auth

import (
	"context"
	servicesv1 "proto/services/core/v1"
	"proto/services/core/v1/corev1connect"

	"web-app-template-core/internal/helper"

	"connectrpc.com/connect"
	errort "github.com/hilmoo/gomox/transport/error"
	"github.com/hilmoo/gomox/transport/validation"
	"github.com/ory/herodot"
)

type connectHandler struct {
	corev1connect.UnimplementedAuthServiceHandler

	svc      Service
	validate *validation.Vld
	isProd   bool
}

var _ corev1connect.AuthServiceHandler = (*connectHandler)(nil)

func NewHandler(svc Service, validate *validation.Vld, isProd bool) corev1connect.AuthServiceHandler {
	return &connectHandler{svc: svc, validate: validate, isProd: isProd}
}

func (h *connectHandler) AuthLogin(
	ctx context.Context,
	req *servicesv1.AuthLoginRequest,
) (*servicesv1.AuthLoginResponse, error) {
	callInfo, errH := handlerCallInfo(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}
	if errH = h.validate.ValidatePayload(req); errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	res, err := h.svc.Login(ctx, LoginInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Meta:     sessionMeta(callInfo),
	})
	if err != nil {
		return nil, toConnectError(ctx, err)
	}

	setSessionCookie(callInfo.ResponseHeader(), res.Token, h.isProd)
	return &servicesv1.AuthLoginResponse{Data: toProtoMeUser(res.User)}, nil
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

	res, err := h.svc.Register(ctx, RegisterInput{
		Email:    req.GetEmail(),
		FullName: req.GetFullName(),
		Password: req.GetPassword(),
		Meta:     sessionMeta(callInfo),
	})
	if err != nil {
		return nil, toConnectError(ctx, err)
	}

	setSessionCookie(callInfo.ResponseHeader(), res.Token, h.isProd)
	return &servicesv1.AuthRegisterResponse{Data: toProtoMeUser(res.User)}, nil
}

func (h *connectHandler) AuthLogout(
	ctx context.Context,
	_ *servicesv1.AuthLogoutRequest,
) (*servicesv1.AuthLogoutResponse, error) {
	callInfo, errH := handlerCallInfo(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}
	// Transport work on the raw token: logout deletes the session the cookie names.
	token, errH := helper.SessionTokenFromHeader(callInfo.RequestHeader())
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	if err := h.svc.Logout(ctx, token); err != nil {
		return nil, toConnectError(ctx, err)
	}

	clearSessionCookie(callInfo.ResponseHeader(), h.isProd)
	return &servicesv1.AuthLogoutResponse{}, nil
}

func (h *connectHandler) AuthMe(
	ctx context.Context,
	_ *servicesv1.AuthMeRequest,
) (*servicesv1.AuthMeResponse, error) {
	userID, errH := helper.UserIDFromContext(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	user, err := h.svc.Me(ctx, userID)
	if err != nil {
		return nil, toConnectError(ctx, err)
	}

	return &servicesv1.AuthMeResponse{Data: toProtoMeUser(user)}, nil
}

func (h *connectHandler) AuthUpdateMe(
	ctx context.Context,
	req *servicesv1.AuthUpdateMeRequest,
) (*servicesv1.AuthUpdateMeResponse, error) {
	userID, errH := helper.UserIDFromContext(ctx)
	if errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}
	if errH = h.validate.ValidatePayload(req); errH != nil {
		return nil, errort.ConnectError(ctx, errH)
	}

	user, err := h.svc.UpdateMe(ctx, userID, UpdateMeInput{
		Email:    req.Email,
		FullName: req.FullName,
	})
	if err != nil {
		return nil, toConnectError(ctx, err)
	}

	return &servicesv1.AuthUpdateMeResponse{Data: toProtoMeUser(user)}, nil
}

func handlerCallInfo(ctx context.Context) (connect.CallInfo, *herodot.DefaultError) {
	callInfo, ok := connect.CallInfoForHandlerContext(ctx)
	if !ok {
		return nil, herodot.ErrInternalServerError.WithReason("missing request metadata")
	}
	return callInfo, nil
}

func sessionMeta(callInfo connect.CallInfo) SessionMeta {
	return SessionMeta{
		IPAddress: peerIP(callInfo.Peer().Addr),
		UserAgent: strPtr(callInfo.RequestHeader().Get("User-Agent")),
	}
}
