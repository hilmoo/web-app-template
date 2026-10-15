import "@typespec/http";
import "@typespec/openapi";
import "@typespec/openapi3";

import "../common/error.tsp";
import "../common/connect.tsp";
import "./model.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace CoreServices;

@tag("Todo")
interface Todo {
  @doc("Create a new todo.")
  @post
  @route("/services.v1.TodoService/TodoCreate")
  create(@body body: TodoModel.CreateRequest):
    | TodoModel.Base
    | CommonError.CreateCommonErrorAuth;

  @doc("Get a todo by its ID.")
  @post
  @route("/services.v1.TodoService/TodoGet")
  get(@body body: TodoModel.Id):
    | TodoModel.Base
    | CommonError.GetCommonErrorAuth;

  @doc("List all todos of the current user.")
  @post
  @route("/services.v1.TodoService/TodoList")
  list(): TodoModel.Base[] | CommonError.GetCommonErrorAuth;

  @doc("Update a todo by its ID.")
  @post
  @route("/services.v1.TodoService/TodoUpdate")
  update(@body body: TodoModel.UpdateRequest):
    | TodoModel.Base
    | CommonError.UpdateCommonErrorAuth;

  @doc("Delete a todo by its ID.")
  @post
  @route("/services.v1.TodoService/TodoDelete")
  delete(@body body: TodoModel.Id):
    | CommonConnect.Empty
    | CommonError.DeleteCommonErrorAuth;
}
