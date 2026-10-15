// AUTO-GENERATED from migration/src/schema by protoforge (embedded drizzle-to-typespec).
// DO NOT EDIT. Run `protoforge stage typespec --only from_drizzle` to regenerate.

import "@typespec/openapi";
import "@typespec/openapi3";

namespace DbSchema;

@format("uuid")
scalar Uuid extends string;

@doc("Database table `sessions`.")
model SessionRow {
  @doc("Unique identifier.")
  @format("uuid")
  @visibility(Lifecycle.Read, Lifecycle.Delete)
  id: string;

  @doc("ID of the user who owns this session.")
  @format("uuid")
  @removeVisibility(Lifecycle.Update)
  userId: string;

  @doc("Hash of the session token.")
  tokenHash: string;

  @doc("User agent of the client that created the session.")
  userAgent?: string;

  @doc("IP address of the client that created the session.")
  ipAddress?: string;

  @doc("When the session expires.")
  expiresAt: utcDateTime;

  @doc("When it was created.")
  @visibility(Lifecycle.Read, Lifecycle.Delete)
  createdAt: utcDateTime;
}

@doc("Database table `todos`.")
model TodoRow {
  @doc("Unique identifier.")
  @format("uuid")
  @visibility(Lifecycle.Read, Lifecycle.Delete)
  id: string;

  @doc("Short title of the todo.")
  @example("Buy milk")
  title: string;

  @doc("Optional longer description.")
  description?: string;

  @doc("Whether the todo has been completed.")
  done: boolean;

  @doc("ID of the user who owns the todo.")
  @format("uuid")
  @removeVisibility(Lifecycle.Update)
  ownerId: string;

  @doc("When it was created.")
  @visibility(Lifecycle.Read, Lifecycle.Delete)
  createdAt: utcDateTime;
}

@doc("Database table `users`.")
model UserRow {
  @doc("Unique identifier.")
  @format("uuid")
  @visibility(Lifecycle.Read, Lifecycle.Delete)
  id: string;

  @doc("Email address of the user.")
  @example("user@example.com")
  email: string;

  @doc("Hash of the user's password. Null for passwordless accounts.")
  passwordHash?: string;

  @doc("Full name of the user.")
  @example("Jane Doe")
  fullName: string;

  @doc("When it was created.")
  @visibility(Lifecycle.Read, Lifecycle.Delete)
  createdAt: utcDateTime;
}
