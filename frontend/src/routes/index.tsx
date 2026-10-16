import { createFileRoute } from "@tanstack/react-router";
import { requireAuth } from "#/features/auth";
import { HomePage } from "#/features/home";

export const Route = createFileRoute("/")({
  beforeLoad: requireAuth,
  component: HomePage,
});
