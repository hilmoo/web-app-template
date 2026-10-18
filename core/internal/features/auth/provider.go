package auth

import (
	"proto/services/v1/servicesv1connect"

	"github.com/hilmoo/gomox/transport/validation"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Pool     *pgxpool.Pool
	Validate *validation.Vld
	Secret   string
	IsProd   bool
}

// Bundle exposes the service too: internal/app passes it to the auth interceptor
// as its helper.SessionResolver.
type Bundle struct {
	Handler servicesv1connect.AuthServiceHandler
	Svc     Service
}

func Provide(d Deps) Bundle {
	repo := NewPostgresRepository(d.Pool)
	svc := NewService(repo, d.Secret)
	return Bundle{Handler: NewHandler(svc, d.Validate, d.IsProd), Svc: svc}
}
