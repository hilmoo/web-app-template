import { createFileRoute } from "@tanstack/react-router";
import { requireAuth } from "#/features/auth";
import { TodosPage } from "#/features/todos";

export const Route = createFileRoute("/")({
  beforeLoad: requireAuth,
  component: TodosPage,
});
