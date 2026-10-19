# Reference Feature: `note` (list, create, delete)

The frontend half of core-pattern's `note` reference feature: a signed-in user's private
notes. Every layer is as small as it can be while still being the real layer. Read this first.
Read `features/auth` second: it adds route guards, cache `set` instead of invalidate, and a
`fetch*` helper for guards.

The contract (`NoteCreate`, `NoteGet`, `NoteList`, `NoteDelete` on `NoteService`) and codegen
come first. See the project-pattern skill. After `task gen:api` you have:

- `NoteService` in `proto/servicesV1` (RPC descriptors for connect-query)
- `Note` and the request/response messages in `proto/modelsV1` / `proto/servicesV1`
- `NoteCreateBody` (zod) in `src/api/schemas/note/note.ts`

Generated names follow the TypeSpec models. Check the generated file before importing.

## Order to write the files

Dependencies point down. Write top to bottom and each file typechecks against the previous ones.

| #   | File                              | One-line job                                                  |
| --- | --------------------------------- | ------------------------------------------------------------- |
| 1   | `api/use-note-list.ts`            | Query hook + `noteListQueryKey`                               |
| 2   | `api/use-note-create.ts`          | Mutation, invalidates the list                                |
| 3   | `api/use-note-delete.ts`          | Mutation, invalidates the list                                |
| 4   | `schemas.ts`                      | `noteCreateSchema` from the orval body + UI messages          |
| 5   | `components/note-create-form.tsx` | Form: zod resolver, error alert, reset on success             |
| 6   | `components/note-list.tsx`        | Renders the list, delete button per row                       |
| 7   | `pages/notes-page.tsx`            | Composes form + list inside `AppShell`                        |
| 8   | `index.ts`                        | Exports `NotesPage` only                                      |
| 9   | `src/routes/notes.tsx`            | `createFileRoute("/notes")` with `requireAuth`                |
| 10  | `schemas.test.ts`                 | Schema accepts valid, rejects empty with the friendly message |

`NoteGet` gets a hook when a page needs it (e.g. `/notes/$id`). Do not write hooks for RPCs no
component calls yet.

## 1. `api/use-note-list.ts`

```ts
import { createConnectQueryKey, useQuery } from "@connectrpc/connect-query";
import { NoteService } from "proto/servicesCoreV1";

const noteList = NoteService.method.noteList;

export const noteListQueryKey = () =>
  createConnectQueryKey({ schema: noteList, cardinality: "finite" });

export function useNoteList() {
  return useQuery(noteList, {}, { select: (res) => res.data });
}
```

The key omits `input`, so it matches every `NoteList` query regardless of arguments. That is
what invalidation wants.

## 2. `api/use-note-create.ts`

```ts
import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { NoteService } from "proto/servicesCoreV1";
import { noteListQueryKey } from "./use-note-list";

export function useNoteCreate() {
  const queryClient = useQueryClient();
  return useMutation(NoteService.method.noteCreate, {
    onSuccess: () => queryClient.invalidateQueries({ queryKey: noteListQueryKey() }),
  });
}
```

## 3. `api/use-note-delete.ts`

Same shape as create, with `NoteService.method.noteDelete`. Even when a hook is this thin it
stays a file of its own (see _Every RPC has an api hook_ in SKILL.md).

## 4. `schemas.ts`

```ts
import * as z from "zod";
import { NoteCreateBody } from "#/api/schemas/note/note";

export const noteCreateSchema = NoteCreateBody.extend({
  title: z.string().trim().min(1, "Title is required"),
});
export type noteCreateSchemaTypes = z.input<typeof noteCreateSchema>;
```

Extend only fields whose generated rule needs a friendlier message or a stricter UX check. Keep
the rest from the generated body, so the form cannot drift from the contract.

## 5. `components/note-create-form.tsx`

```tsx
import { Alert, Button, Stack, TextInput, Textarea } from "@mantine/core";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { getErrorMessage } from "#/shared/lib/errors";
import { useNoteCreate } from "../api/use-note-create";
import { noteCreateSchema, type noteCreateSchemaTypes } from "../schemas";

export function NoteCreateForm() {
  const create = useNoteCreate();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<noteCreateSchemaTypes>({
    resolver: zodResolver(noteCreateSchema),
    defaultValues: { title: "", body: "" },
  });

  const onSubmit = handleSubmit((values) => create.mutate(values, { onSuccess: () => reset() }));

  return (
    <form onSubmit={onSubmit} noValidate>
      <Stack>
        {create.isError && (
          <Alert color="red" variant="light">
            {getErrorMessage(create.error)}
          </Alert>
        )}
        <TextInput label="Title" error={errors.title?.message} {...register("title")} />
        <Textarea
          label="Body"
          autosize
          minRows={3}
          error={errors.body?.message}
          {...register("body")}
        />
        <Button type="submit" loading={create.isPending}>
          Add note
        </Button>
      </Stack>
    </form>
  );
}
```

## 6. `components/note-list.tsx`

```tsx
import { Alert, Button, Card, Group, Loader, Stack, Text } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { getErrorMessage } from "#/shared/lib/errors";
import { useNoteDelete } from "../api/use-note-delete";
import { useNoteList } from "../api/use-note-list";

export function NoteList() {
  const { data: notes, isPending, error } = useNoteList();
  const remove = useNoteDelete();

  if (isPending) {
    return <Loader />;
  }
  if (error) {
    return (
      <Alert color="red" variant="light">
        {getErrorMessage(error)}
      </Alert>
    );
  }
  if (notes.length === 0) {
    return <Text c="dimmed">No notes yet.</Text>;
  }

  return (
    <Stack>
      {notes.map((note) => (
        <Card key={note.id} withBorder>
          <Group justify="space-between">
            <Text fw={500}>{note.title}</Text>
            <Button
              color="red"
              variant="subtle"
              loading={remove.isPending && remove.variables?.noteId === note.id}
              onClick={() =>
                remove.mutate(
                  { noteId: note.id },
                  {
                    onError: (err) =>
                      notifications.show({ message: getErrorMessage(err), color: "red" }),
                  },
                )
              }
            >
              Delete
            </Button>
          </Group>
          {note.body && <Text size="sm">{note.body}</Text>}
        </Card>
      ))}
    </Stack>
  );
}
```

Proto `Note` messages go straight into JSX. There is no mapper and no view model. If a
component needs a typed prop, use `import type { Note } from "proto/modelsV1"`.

## 7. `pages/notes-page.tsx`

```tsx
import { Container, Stack, Title } from "@mantine/core";
import { AppShell } from "#/features/layout";
import { NoteCreateForm } from "../components/note-create-form";
import { NoteList } from "../components/note-list";

export function NotesPage() {
  return (
    <AppShell>
      <Container size="sm" py="xl">
        <Stack>
          <Title order={2}>Notes</Title>
          <NoteCreateForm />
          <NoteList />
        </Stack>
      </Container>
    </AppShell>
  );
}
```

The page composes and (when needed) navigates. In auth, `LoginPage` passes
`onSuccess={() => navigate({ to: "/" })}` into `LoginForm`, so the form stays reusable and
navigation is decided in one place.

## 8. `index.ts`

```ts
// Public API of the note feature. Other code imports from "#/features/note" only.
export { NotesPage } from "./pages/notes-page";
```

Export hooks or components only when another feature or route actually uses them. Auth exports
`useAuthMe` because `layout` and `home` show the user.

## 9. `src/routes/notes.tsx`

```tsx
import { createFileRoute } from "@tanstack/react-router";
import { requireAuth } from "#/features/auth";
import { NotesPage } from "#/features/note";

export const Route = createFileRoute("/notes")({
  beforeLoad: requireAuth,
  component: NotesPage,
});
```

This is the frontend's `register(...)` line: one file, no logic. `routeTree.gen.ts` updates
when the dev server is running, or with `pnpm --filter frontend generate-routes`. Commit it.
Core's procedures require auth by default. On the frontend you opt in per route with
`requireAuth`, so add it to every signed-in route.

## 10. `schemas.test.ts`

```ts
import { describe, expect, it } from "vitest";
import { noteCreateSchema } from "./schemas";

describe("noteCreateSchema", () => {
  it("accepts a title", () => {
    expect(noteCreateSchema.safeParse({ title: "Groceries", body: "" }).success).toBe(true);
  });

  it("rejects a blank title with a friendly message", () => {
    const result = noteCreateSchema.safeParse({ title: "  ", body: "" });
    expect(result.success).toBe(false);
    expect(result.error?.issues[0]?.message).toBe("Title is required");
  });
});
```
