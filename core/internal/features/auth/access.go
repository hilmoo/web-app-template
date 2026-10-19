package auth

import (
	"proto/services/core/v1/corev1connect"

	"web-app-template-core/internal/helper"
)

// Access lists the auth procedures that differ from the default (authenticated).
//
//nolint:gochecknoglobals // static policy table
var Access = helper.AccessPolicy{
	corev1connect.AuthServiceAuthLoginProcedure:    helper.AccessPublic,
	corev1connect.AuthServiceAuthRegisterProcedure: helper.AccessPublic,
}
