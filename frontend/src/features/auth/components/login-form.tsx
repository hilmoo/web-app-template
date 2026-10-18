import { Alert, Button, PasswordInput, Stack, TextInput } from "@mantine/core";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { getErrorMessage } from "#/shared/lib/errors";
import { useAuthLogin } from "../api/use-auth-login";
import { loginSchema, type loginSchemaTypes } from "../schemas";

export function LoginForm({ onSuccess }: { onSuccess: () => void }) {
  const login = useAuthLogin();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<loginSchemaTypes>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });

  const onSubmit = handleSubmit((values) => login.mutate(values, { onSuccess }));

  return (
    <form onSubmit={onSubmit} noValidate>
      <Stack>
        {login.isError && (
          <Alert color="red" variant="light">
            {getErrorMessage(login.error)}
          </Alert>
        )}
        <TextInput
          label="Email"
          type="email"
          autoComplete="email"
          error={errors.email?.message}
          {...register("email")}
        />
        <PasswordInput
          label="Password"
          autoComplete="current-password"
          error={errors.password?.message}
          {...register("password")}
        />
        <Button type="submit" loading={login.isPending}>
          Sign in
        </Button>
      </Stack>
    </form>
  );
}
