import { Alert, Center, Container, Loader, Stack, Text, Title } from "@mantine/core";
import { AppShell } from "#/shared/components/app-shell";
import { getErrorMessage } from "#/shared/lib/errors";
import { useTodos } from "../api/use-todos";
import { CreateTodoForm } from "../components/create-todo-form";
import { TodoItem } from "../components/todo-item";

export function TodosPage() {
  const { data: todos, isPending, error } = useTodos();

  return (
    <AppShell>
      <Container size="sm" py="xl">
        <Title order={2} mb="md">
          Todos
        </Title>
        <Stack>
          <CreateTodoForm />

          {error && (
            <Alert color="red" variant="light">
              {getErrorMessage(error)}
            </Alert>
          )}
          {isPending && (
            <Center py="xl">
              <Loader />
            </Center>
          )}
          {todos?.length === 0 && (
            <Text c="dimmed" ta="center" py="xl">
              Nothing to do yet.
            </Text>
          )}
          {todos?.map((todo) => (
            <TodoItem key={todo.id} todo={todo} />
          ))}
        </Stack>
      </Container>
    </AppShell>
  );
}
