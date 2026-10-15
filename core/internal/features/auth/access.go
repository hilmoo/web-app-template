package auth

import (
	"proto/services/v1/servicesv1connect"

	"web-app-template-core/internal/helper"
)

// Access lists the auth procedures that differ from the default (authenticated).
//
//nolint:gochecknoglobals // static policy table
var Access = helper.AccessPolicy{
	servicesv1connect.AuthServiceAuthLoginProcedure:    helper.AccessPublic,
	servicesv1connect.AuthServiceAuthRegisterProcedure: helper.AccessPublic,
}
