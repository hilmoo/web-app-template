import { createConnectQueryKey } from "@connectrpc/connect-query";
import { TodoService } from "proto/servicesV1";

export const todoListQueryKey = () =>
  createConnectQueryKey({
    schema: TodoService.method.todoList,
    input: {},
    cardinality: "finite",
  });
