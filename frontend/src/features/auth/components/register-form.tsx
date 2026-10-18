import { Alert, Button, PasswordInput, Stack, TextInput } from "@mantine/core";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { getErrorMessage } from "#/shared/lib/errors";
import { useAuthRegister } from "../api/use-auth-register";
import { registerSchema, type registerSchemaTypes } from "../schemas";

export function RegisterForm({ onSuccess }: { onSuccess: () => void }) {
  const registerUser = useAuthRegister();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<registerSchemaTypes>({
    resolver: zodResolver(registerSchema),
    defaultValues: { fullName: "", email: "", password: "" },
  });

  const onSubmit = handleSubmit((values) => registerUser.mutate(values, { onSuccess }));

  return (
    <form onSubmit={onSubmit} noValidate>
      <Stack>
        {registerUser.isError && (
          <Alert color="red" variant="light">
            {getErrorMessage(registerUser.error)}
          </Alert>
        )}
        <TextInput
          label="Full name"
          autoComplete="name"
          error={errors.fullName?.message}
          {...register("fullName")}
        />
        <TextInput
          label="Email"
          type="email"
          autoComplete="email"
          error={errors.email?.message}
          {...register("email")}
        />
        <PasswordInput
          label="Password"
          autoComplete="new-password"
          error={errors.password?.message}
          {...register("password")}
        />
        <Button type="submit" loading={registerUser.isPending}>
          Create account
        </Button>
      </Stack>
    </form>
  );
}
