package app

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"proto/services/v1/servicesv1connect"

	"web-app-template-core/internal/features/auth"
	"web-app-template-core/internal/features/health"
	"web-app-template-core/internal/helper"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	mlog "github.com/hilmoo/gomox/transport/middleware/log"
)

func Handler(arg helper.HandlerParams) http.Handler {
	api := http.NewServeMux()
	RegisterConnect(arg, api)

	healthMux := http.NewServeMux()
	health.NewHTTP().Register(healthMux)

	web := http.NewServeMux()
	registerFrontend(web)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api))
	mux.Handle("GET /api/health", healthMux)
	mux.Handle("/", web)

	return mux
}

// AccessInternalOnly lists procedures callable only with the service secret.
// Procedures that are not listed (here or in a feature's own policy) require authentication.
var AccessInternalOnly = helper.AccessPolicy{} //nolint:gochecknoglobals // static policy table

func RegisterConnect(handleArg helper.HandlerParams, mux *http.ServeMux) []string {
	authBundle := auth.Provide(auth.Deps{
		Pool:     handleArg.Pool,
		Validate: handleArg.Validate,
		Secret:   handleArg.Config.Secret,
		IsProd:   handleArg.Config.IsProd,
	})

	opts := []connect.HandlerOption{
		connect.WithInterceptors(
			mlog.ConnectLogger(handleArg.Logger),
			helper.NewAuthInterceptor(
				handleArg,
				authBundle.Svc,
				auth.Access,
				AccessInternalOnly,
			),
		),
	}
	if handleArg.Config.IsProd {
		opts = append(
			opts,
			connect.WithRecover(func(ctx context.Context, spec connect.Spec, _ http.Header, panicValue any) error {
				handleArg.Logger.ErrorContext(
					ctx,
					"connect handler panic",
					"procedure",
					spec.Procedure,
					"panic",
					panicValue,
				)
				return connect.NewError(connect.CodeInternal, errors.New("internal error"))
			}),
		)
	}

	var serviceNames []string
	register := func(path string, h http.Handler) {
		mux.Handle(path, h)
		serviceNames = append(serviceNames, strings.Trim(path, "/"))
	}

	register(servicesv1connect.NewAuthServiceHandler(authBundle.Handler, opts...))

	reflector := grpcreflect.NewStaticReflector(serviceNames...)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	return serviceNames
}
