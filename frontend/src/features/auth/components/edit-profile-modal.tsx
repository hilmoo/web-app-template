import { Alert, Button, Modal, Stack, TextInput } from "@mantine/core";
import { zodResolver } from "@hookform/resolvers/zod";
import { notifications } from "@mantine/notifications";
import { useForm } from "react-hook-form";
import { getErrorMessage } from "#/shared/lib/errors";
import { useAuthMe } from "../api/use-auth-me";
import { useAuthUpdateMe } from "../api/use-auth-update-me";
import { updateMeSchema, type updateMeSchemaTypes } from "../schemas";

function ProfileForm({ onClose }: { onClose: () => void }) {
  const { data: user } = useAuthMe();
  const update = useAuthUpdateMe();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<updateMeSchemaTypes>({
    resolver: zodResolver(updateMeSchema),
    defaultValues: { fullName: user?.fullName ?? "", email: user?.email ?? "" },
  });

  const onSubmit = handleSubmit((values) =>
    update.mutate(values, {
      onSuccess: () => {
        notifications.show({ message: "Profile updated", color: "green" });
        onClose();
      },
    }),
  );

  return (
    <form onSubmit={onSubmit} noValidate>
      <Stack>
        {update.isError && (
          <Alert color="red" variant="light">
            {getErrorMessage(update.error)}
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
        <Button type="submit" loading={update.isPending}>
          Save
        </Button>
      </Stack>
    </form>
  );
}

export function EditProfileModal({ opened, onClose }: { opened: boolean; onClose: () => void }) {
  return (
    <Modal opened={opened} onClose={onClose} title="Edit profile">
      {opened && <ProfileForm onClose={onClose} />}
    </Modal>
  );
}
