import { redirect } from "@tanstack/react-router";
import type { RouterContext } from "#/routes/__root";
import { fetchMe } from "./api/me";

/** Route `beforeLoad` guard: send anonymous visitors to /login. */
export async function requireAuth({ context }: { context: RouterContext }) {
  const user = await fetchMe(context.queryClient, context.transport);
  if (!user) {
    throw redirect({ to: "/login" });
  }
  return { user };
}

/** Route `beforeLoad` guard: keep signed-in users away from /login and /register. */
export async function redirectIfAuthed({ context }: { context: RouterContext }) {
  const user = await fetchMe(context.queryClient, context.transport);
  if (user) {
    throw redirect({ to: "/" });
  }
}
