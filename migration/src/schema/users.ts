import { pgTable, text, timestamp, uuid } from "drizzle-orm/pg-core";

export const users = pgTable("users", {
  // Unique identifier.
  id: uuid("id").primaryKey().defaultRandom(),
  // Email address of the user.
  // @example "user@example.com"
  email: text("email").notNull().unique(),
  // Hash of the user's password. Null for passwordless accounts.
  passwordHash: text("password_hash"),
  // Full name of the user.
  // @example "Jane Doe"
  fullName: text("full_name").notNull(),

  // When it was created.
  createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
});
