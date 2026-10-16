import { Container, Text, Title } from "@mantine/core";
import { useMe } from "#/features/auth";
import { AppShell } from "#/shared/components/app-shell";

export function HomePage() {
  const { data: user } = useMe();

  return (
    <AppShell>
      <Container size="sm" py="xl">
        <Title order={2} mb="xs">
          Welcome{user?.fullName ? `, ${user.fullName}` : ""}
        </Title>
        <Text c="dimmed">Your app starts here.</Text>
      </Container>
    </AppShell>
  );
}
