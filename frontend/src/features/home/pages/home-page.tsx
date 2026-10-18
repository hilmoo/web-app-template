import { Container, Text, Title } from "@mantine/core";
import { useAuthMe } from "#/features/auth";
import { AppShell } from "#/features/layout";

export function HomePage() {
  const { data: user } = useAuthMe();

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
