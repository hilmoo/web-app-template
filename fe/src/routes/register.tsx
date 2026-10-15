import { createFileRoute } from "@tanstack/react-router";
import { RegisterPage, redirectIfAuthed } from "#/features/auth";

export const Route = createFileRoute("/register")({
  beforeLoad: redirectIfAuthed,
  component: RegisterPage,
});
