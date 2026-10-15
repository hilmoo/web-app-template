import { useQuery } from "@connectrpc/connect-query";
import { TodoService } from "proto/servicesV1";

export function useTodos() {
  return useQuery(TodoService.method.todoList, {}, { select: (res) => res.data });
}
