import { boolean, index, pgTable, text, timestamp, uuid } from "drizzle-orm/pg-core";
import { users } from "./users";

export const todos = pgTable(
  "todos",
  {
    // Unique identifier.
    id: uuid("id").primaryKey().defaultRandom(),

    // Short title of the todo.
    // @example "Buy milk"
    title: text("title").notNull(),
    // Optional longer description.
    description: text("description"),
    // Whether the todo has been completed.
    done: boolean("done").notNull().default(false),
    // ID of the user who owns the todo.
    ownerId: uuid("owner_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),

    // When it was created.
    createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  },
  (table) => [index("idx_todos_owner_id").on(table.ownerId)],
);
