import { Alert, Button, Group, Stack, TextInput } from "@mantine/core";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { getErrorMessage } from "#/shared/lib/errors";
import { useCreateTodo } from "../api/use-create-todo";
import { createTodoSchema, type createTodoSchemaTypes } from "../schemas";

export function CreateTodoForm() {
  const create = useCreateTodo();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<createTodoSchemaTypes>({
    resolver: zodResolver(createTodoSchema),
    defaultValues: { title: "", description: "" },
  });

  const onSubmit = handleSubmit((values) =>
    create.mutate(
      { title: values.title, description: values.description || undefined },
      { onSuccess: () => reset() },
    ),
  );

  return (
    <form onSubmit={onSubmit} noValidate>
      <Stack gap="xs">
        {create.isError && (
          <Alert color="red" variant="light">
            {getErrorMessage(create.error)}
          </Alert>
        )}
        <Group align="flex-start" wrap="nowrap">
          <TextInput
            placeholder="What needs to be done?"
            style={{ flex: 1 }}
            error={errors.title?.message}
            {...register("title")}
          />
          <Button type="submit" loading={create.isPending}>
            Add
          </Button>
        </Group>
      </Stack>
    </form>
  );
}
