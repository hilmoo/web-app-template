package todo

import (
	"context"
	servicesv1 "proto/services/v1"
	"proto/services/v1/servicesv1connect"

	db "web-app-template-core/internal/gen/sqlc"
	"web-app-template-core/internal/helper"
	"web-app-template-core/internal/x/typex"

	"github.com/google/uuid"
	"github.com/hilmoo/gomox/transport/validation"
	"github.com/ory/herodot"
)

type connectHandler struct {
	servicesv1connect.UnimplementedTodoServiceHandler

	validate *validation.Vld
	queries  *db.Queries
}

var _ servicesv1connect.TodoServiceHandler = (*connectHandler)(nil)

func NewConnect(arg helper.HandlerParams) servicesv1connect.TodoServiceHandler {
	return &connectHandler{
		validate: arg.Validate,
		queries:  arg.Queries,
	}
}

func (h *connectHandler) TodoCreate(
	ctx context.Context,
	req *servicesv1.TodoCreateRequest,
) (*servicesv1.TodoCreateResponse, error) {
	return helper.WithUserID(ctx, func(userID uuid.UUID) (*servicesv1.TodoCreateResponse, *herodot.DefaultError) {
		if errH := h.validate.ValidatePayload(req); errH != nil {
			return nil, errH
		}

		todo, errH := todoCreateService(ctx, h.queries, db.TodoCreateParams{
			Title:       req.GetTitle(),
			Description: req.Description,
			OwnerID:     userID,
		})
		if errH != nil {
			return nil, errH
		}

		return &servicesv1.TodoCreateResponse{Data: toTodo(todo)}, nil
	})
}

func (h *connectHandler) TodoDelete(
	ctx context.Context,
	req *servicesv1.TodoDeleteRequest,
) (*servicesv1.TodoDeleteResponse, error) {
	return helper.WithUserID(ctx, func(userID uuid.UUID) (*servicesv1.TodoDeleteResponse, *herodot.DefaultError) {
		id, errH := helper.StrToUUIDHTTP(req.GetTodoId())
		if errH != nil {
			return nil, errH
		}

		if errH = todoDeleteService(ctx, h.queries, db.TodoDeleteParams{ID: id, OwnerID: userID}); errH != nil {
			return nil, errH
		}

		return &servicesv1.TodoDeleteResponse{}, nil
	})
}

func (h *connectHandler) TodoGet(
	ctx context.Context,
	req *servicesv1.TodoGetRequest,
) (*servicesv1.TodoGetResponse, error) {
	return helper.WithUserID(ctx, func(userID uuid.UUID) (*servicesv1.TodoGetResponse, *herodot.DefaultError) {
		id, errH := helper.StrToUUIDHTTP(req.GetTodoId())
		if errH != nil {
			return nil, errH
		}

		todo, errH := todoGetService(ctx, h.queries, db.TodoGetByIdParams{ID: id, OwnerID: userID})
		if errH != nil {
			return nil, errH
		}

		return &servicesv1.TodoGetResponse{Data: toTodo(todo)}, nil
	})
}

func (h *connectHandler) TodoList(
	ctx context.Context,
	_ *servicesv1.TodoListRequest,
) (*servicesv1.TodoListResponse, error) {
	return helper.WithUserID(ctx, func(userID uuid.UUID) (*servicesv1.TodoListResponse, *herodot.DefaultError) {
		todos, errH := todoListService(ctx, h.queries, userID)
		if errH != nil {
			return nil, errH
		}

		return &servicesv1.TodoListResponse{Data: typex.MapSlice(todos, toTodo)}, nil
	})
}

func (h *connectHandler) TodoUpdate(
	ctx context.Context,
	req *servicesv1.TodoUpdateRequest,
) (*servicesv1.TodoUpdateResponse, error) {
	return helper.WithUserID(ctx, func(userID uuid.UUID) (*servicesv1.TodoUpdateResponse, *herodot.DefaultError) {
		id, errH := helper.StrToUUIDHTTP(req.GetTodoId())
		if errH != nil {
			return nil, errH
		}

		todo, errH := todoUpdateService(ctx, h.queries, db.TodoUpdateParams{
			Title:       req.Title,
			Description: req.Description,
			Done:        req.Done,
			ID:          id,
			OwnerID:     userID,
		})
		if errH != nil {
			return nil, errH
		}

		return &servicesv1.TodoUpdateResponse{Data: toTodo(todo)}, nil
	})
}
