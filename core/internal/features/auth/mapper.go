package auth

import (
	modelsv1 "proto/models/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoMeUser(u User) *modelsv1.AuthMeUser {
	return &modelsv1.AuthMeUser{
		Id:        u.ID.String(),
		Email:     u.Email,
		FullName:  u.FullName,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}
}
