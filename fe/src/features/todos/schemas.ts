import * as z from "zod";
import { TodoCreateBody } from "#/api/schemas/todo/todo";

export const createTodoSchema = TodoCreateBody.extend({
  title: z.string().trim().min(1, "Title is required"),
  description: z.string().trim(),
});
export type createTodoSchemaTypes = z.input<typeof createTodoSchema>;
