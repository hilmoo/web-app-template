import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { TodoService } from "proto/servicesV1";
import { todoListQueryKey } from "./keys";

export function useCreateTodo() {
  const queryClient = useQueryClient();
  return useMutation(TodoService.method.todoCreate, {
    onSuccess: () => queryClient.invalidateQueries({ queryKey: todoListQueryKey() }),
  });
}
