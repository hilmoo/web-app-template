# Forms & Errors

Core validates every request and returns Connect errors (see core-pattern's
`references/error-flow.md` and `references/response-shape.md`). The frontend validates only for
UX, then shows whatever core says, through one helper.

## Where validation lives

| Check | Where | Why |
|---|---|---|
| Field shape (email, uuid, required by contract) | Generated `src/api/schemas/<tag>/<tag>.ts` (orval) | Same source as core's proto `validate` tags, so it cannot drift |
| Friendly message, trimming, stricter UX rule | `features/<feature>/schemas.ts` via `.extend({...})` | UI wording is a frontend concern |
| Business rules (uniqueness, ownership, quotas) | core service only | The frontend shows the resulting error and never re-implements the rule |

Never write a zod object from scratch for an RPC body. Start from the generated `<Feature><Op>Body`
and extend it. Shared field rules within a feature (`requiredEmail`) are local `const`s in
`schemas.ts` (see `features/auth/schemas.ts`).

Name schemas after the RPC op (`loginSchema`, `noteCreateSchema`) and export the input type
`z.input<typeof schema>` as `<name>SchemaTypes`. Use `z.input` because the form holds
pre-transform values.

## Form pattern

```tsx
const mutation = useNoteCreate();
const { register, handleSubmit, reset, formState: { errors } } = useForm<noteCreateSchemaTypes>({
  resolver: zodResolver(noteCreateSchema),
  defaultValues: { title: "", body: "" },     // every field, so inputs stay controlled
});
const onSubmit = handleSubmit((values) => mutation.mutate(values, { onSuccess }));

<form onSubmit={onSubmit} noValidate>          // noValidate: zod owns validation, not the browser
  <Stack>
    {mutation.isError && <Alert color="red" variant="light">{getErrorMessage(mutation.error)}</Alert>}
    <TextInput label="Title" error={errors.title?.message} {...register("title")} />
    <Button type="submit" loading={mutation.isPending}>Save</Button>
  </Stack>
</form>
```

- Field errors (client-side) go to the input's `error` prop. Server errors go to one `<Alert>`
  at the top of the form.
- `loading={mutation.isPending}` on the submit button, never a separate `useState` flag.
- `onSuccess` from the page (navigate) or the component (reset, close modal, notification) is
  passed per call in `mutate(values, { onSuccess })`. Cache work stays in the api hook's own
  `onSuccess`, which runs first.
- Forms that load existing values (edit modal) render the form only while open
  (`{opened && <ProfileForm />}`), so `defaultValues` are read fresh each time. See
  `features/auth/components/edit-profile-modal.tsx`.

## Showing errors

`getErrorMessage(error)` in `src/shared/lib/errors.ts` is the only code that reads Connect error
details. It picks, in order:

1. `google.rpc.BadRequest.fieldViolations` → `"field: description"` lines (core's `ValidatePayload`)
2. `google.rpc.ErrorInfo.reason` → the user-safe message core set with `WithReason(...)`
   (e.g. "invalid email or password")
3. `ConnectError.rawMessage`, then a generic fallback

| Situation | Display |
|---|---|
| Form submit fails | `<Alert color="red" variant="light">` above the fields |
| In-place action fails (delete, toggle) | `notifications.show({ message: getErrorMessage(err), color: "red" })` in the per-call `onError` |
| In-place action succeeds and nothing visible changes | `notifications.show({ message, color: "green" })` |
| Query fails | `<Alert>` in place of the content |
| `Unauthenticated` from a guard | `requireAuth` redirects to `/login` (see `fetchAuthMe`) |

Do not branch on error codes in components. If a flow genuinely depends on a code (as the auth
guard treats `Unauthenticated` / `PermissionDenied` as "signed out"), do it in the `api/` or
`guards.ts` layer with `ConnectError.from(error).code`.

If core needs a new user-facing message, add a reason in the core feature's `toAppError`. Don't
map codes to strings on the frontend.

## Testing

- `schemas.test.ts` per feature: valid input passes, each friendly message appears for its
  failure. These are cheap and catch drift after regenerating orval schemas.
- Pure helpers (`shared/lib/*`) get `*.test.ts` next to them.
- Run with `task frontend:test` (vitest).
