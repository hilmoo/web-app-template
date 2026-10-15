import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { TodoService } from "proto/servicesV1";
import { todoListQueryKey } from "./keys";

export function useDeleteTodo() {
  const queryClient = useQueryClient();
  return useMutation(TodoService.method.todoDelete, {
    onSuccess: () => queryClient.invalidateQueries({ queryKey: todoListQueryKey() }),
  });
}
