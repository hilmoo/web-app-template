package todo

import (
	modelsv1 "proto/models/v1"

	"google.golang.org/protobuf/types/known/timestamppb"

	db "web-app-template-core/internal/gen/sqlc"
)

func toTodo(t db.Todo) *modelsv1.TodoModelBase {
	return &modelsv1.TodoModelBase{
		Id:          t.ID.String(),
		Title:       t.Title,
		Description: t.Description,
		Done:        t.Done,
		OwnerId:     t.OwnerID.String(),
		CreatedAt:   timestamppb.New(t.CreatedAt),
	}
}
