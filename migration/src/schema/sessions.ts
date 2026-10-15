import { index, pgTable, text, timestamp, uuid } from "drizzle-orm/pg-core";
import { users } from "./users";

export const sessions = pgTable(
  "sessions",
  {
    // Unique identifier.
    id: uuid("id").primaryKey().defaultRandom(),
    // ID of the user who owns this session.
    userId: uuid("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),

    // Hash of the session token.
    tokenHash: text("token_hash").notNull().unique(),
    // User agent of the client that created the session.
    userAgent: text("user_agent"),
    // IP address of the client that created the session.
    ipAddress: text("ip_address"),

    // When the session expires.
    expiresAt: timestamp("expires_at", { withTimezone: true }).notNull(),
    // When it was created.
    createdAt: timestamp("created_at", { withTimezone: true }).notNull().defaultNow(),
  },
  (table) => [
    index("idx_sessions_user_id").on(table.userId),
    index("idx_sessions_expires_at").on(table.expiresAt),
  ],
);
