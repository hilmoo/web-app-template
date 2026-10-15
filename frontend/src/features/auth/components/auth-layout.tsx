import type { ReactNode } from "react";
import { Anchor, Container, Paper, Text, Title } from "@mantine/core";
import { Link } from "@tanstack/react-router";

interface AuthLayoutProps {
  title: string;
  footerText: string;
  footerLink: { to: "/login" | "/register"; label: string };
  children: ReactNode;
}

export function AuthLayout({ title, footerText, footerLink, children }: AuthLayoutProps) {
  return (
    <Container size={420} my={80}>
      <Title ta="center">{title}</Title>
      <Text c="dimmed" size="sm" ta="center" mt={5}>
        {footerText}{" "}
        <Anchor size="sm" component={Link} to={footerLink.to}>
          {footerLink.label}
        </Anchor>
      </Text>
      <Paper withBorder shadow="sm" p={30} mt={30} radius="md">
        {children}
      </Paper>
    </Container>
  );
}
