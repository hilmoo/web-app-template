import { createFileRoute } from "@tanstack/react-router";
import { LoginPage, redirectIfAuthed } from "#/features/auth";

export const Route = createFileRoute("/login")({
  beforeLoad: redirectIfAuthed,
  component: LoginPage,
});
