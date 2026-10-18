import { useState, type ReactNode } from "react";
import { Avatar, Box, Container, Group, Menu, Text, UnstyledButton } from "@mantine/core";
import { Link, useNavigate } from "@tanstack/react-router";
import { EditProfileModal, useAuthLogout, useAuthMe } from "#/features/auth";

function initials(name?: string, email?: string) {
  const source = name?.trim() || email || "?";
  return source
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}

export function AppShell({ children }: { children: ReactNode }) {
  const navigate = useNavigate();
  const { data: user } = useAuthMe();
  const logout = useAuthLogout();
  const [editingProfile, setEditingProfile] = useState(false);

  return (
    <Box mih="100vh">
      <Box
        component="header"
        h={60}
        px="md"
        style={{
          position: "sticky",
          top: 0,
          zIndex: 100,
          borderBottom: "1px solid var(--mantine-color-default-border)",
          background: "var(--mantine-color-body)",
        }}
      >
        <Container size="xl" h="100%" px={0}>
          <Group h="100%" justify="space-between">
            <Link to="/" style={{ textDecoration: "none", color: "inherit" }}>
              <Group gap="xs">
                <Avatar radius="md" color="blue" variant="filled" size={30}>
                  W
                </Avatar>
                <Text fw={700} size="lg">
                  Web App
                </Text>
              </Group>
            </Link>

            <Menu position="bottom-end" width={220} shadow="md">
              <Menu.Target>
                <UnstyledButton aria-label="Account menu">
                  <Group gap="xs">
                    <Box visibleFrom="sm" ta="right">
                      <Text size="sm" fw={500} lh={1.2}>
                        {user?.fullName}
                      </Text>
                      <Text size="xs" c="dimmed" lh={1.2}>
                        {user?.email}
                      </Text>
                    </Box>
                    <Avatar radius="xl" color="blue">
                      {initials(user?.fullName, user?.email)}
                    </Avatar>
                  </Group>
                </UnstyledButton>
              </Menu.Target>
              <Menu.Dropdown>
                <Menu.Label>{user?.email}</Menu.Label>
                <Menu.Item onClick={() => setEditingProfile(true)}>Edit profile</Menu.Item>
                <Menu.Divider />
                <Menu.Item
                  color="red"
                  disabled={logout.isPending}
                  onClick={() => logout.mutate({}, { onSuccess: () => navigate({ to: "/login" }) })}
                >
                  Sign out
                </Menu.Item>
              </Menu.Dropdown>
            </Menu>
          </Group>
        </Container>
      </Box>
      <main>{children}</main>
      <EditProfileModal opened={editingProfile} onClose={() => setEditingProfile(false)} />
    </Box>
  );
}
