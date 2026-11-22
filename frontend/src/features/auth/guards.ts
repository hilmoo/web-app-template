import { redirect } from "@tanstack/react-router";
import type { RouterContext } from "#/routes/__root";
import { fetchAuthMe } from "./api/use-auth-me";

export async function requireAuth({ context }: { context: RouterContext }) {
  const user = await fetchAuthMe(context.queryClient, context.transport);
  if (!user) {
    throw redirect({ to: "/login" });
  }
  return { user };
}

export async function redirectIfAuthed({ context }: { context: RouterContext }) {
  const user = await fetchAuthMe(context.queryClient, context.transport);
  if (user) {
    throw redirect({ to: "/" });
  }
}
