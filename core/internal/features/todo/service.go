package todo

import (
	"context"

	db "web-app-template-core/internal/gen/sqlc"
	"web-app-template-core/internal/helper"

	"github.com/google/uuid"
	"github.com/ory/herodot"
)

const resource = "todo"

func todoCreateService(
	ctx context.Context,
	queries *db.Queries,
	arg db.TodoCreateParams,
) (db.Todo, *herodot.DefaultError) {
	todo, err := queries.TodoCreate(ctx, arg)
	if err != nil {
		return db.Todo{}, helper.MapDBErrorHTTP(err, resource)
	}
	return todo, nil
}

func todoGetService(
	ctx context.Context,
	queries *db.Queries,
	arg db.TodoGetByIdParams,
) (db.Todo, *herodot.DefaultError) {
	todo, err := queries.TodoGetById(ctx, arg)
	if err != nil {
		return db.Todo{}, helper.MapDBErrorHTTP(err, resource)
	}
	return todo, nil
}

func todoListService(
	ctx context.Context,
	queries *db.Queries,
	userID uuid.UUID,
) ([]db.Todo, *herodot.DefaultError) {
	todos, err := queries.TodoList(ctx, userID)
	if err != nil {
		return nil, helper.MapDBErrorHTTP(err, resource)
	}
	return todos, nil
}

func todoUpdateService(
	ctx context.Context,
	queries *db.Queries,
	arg db.TodoUpdateParams,
) (db.Todo, *herodot.DefaultError) {
	todo, err := queries.TodoUpdate(ctx, arg)
	if err != nil {
		return db.Todo{}, helper.MapDBErrorHTTP(err, resource)
	}
	return todo, nil
}

func todoDeleteService(
	ctx context.Context,
	queries *db.Queries,
	arg db.TodoDeleteParams,
) *herodot.DefaultError {
	rows, err := queries.TodoDelete(ctx, arg)
	if err != nil {
		return helper.MapDBErrorHTTP(err, resource)
	}
	if rows == 0 {
		return herodot.ErrNotFound.WithReason(resource + " not found")
	}
	return nil
}
