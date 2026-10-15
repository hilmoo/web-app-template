import { ActionIcon, Checkbox, Group, Paper, Text } from "@mantine/core";
import type { TodoModelBase } from "proto/modelsV1";
import { useDeleteTodo } from "../api/use-delete-todo";
import { useUpdateTodo } from "../api/use-update-todo";

export function TodoItem({ todo }: { todo: TodoModelBase }) {
  const update = useUpdateTodo();
  const del = useDeleteTodo();

  return (
    <Paper withBorder radius="md" p="sm">
      <Group justify="space-between" wrap="nowrap">
        <Checkbox
          checked={todo.done}
          disabled={update.isPending}
          onChange={(e) => update.mutate({ todoId: todo.id, done: e.currentTarget.checked })}
          label={
            <Text td={todo.done ? "line-through" : undefined} c={todo.done ? "dimmed" : undefined}>
              {todo.title}
            </Text>
          }
        />
        <ActionIcon
          variant="subtle"
          color="red"
          aria-label="Delete todo"
          loading={del.isPending}
          onClick={() => del.mutate({ todoId: todo.id })}
        >
          ×
        </ActionIcon>
      </Group>
    </Paper>
  );
}
