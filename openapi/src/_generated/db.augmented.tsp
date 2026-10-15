// Hand-written layer on top of the generated db.tsp: visibility rules and formats
// that cannot be expressed in the drizzle schema.
import "./db.tsp";

namespace DbSchema;

// Server-assigned: set from the authenticated user, never by the client.
@@visibility(DbSchema.TodoRow.ownerId, Lifecycle.Read, Lifecycle.Delete);

@@format(DbSchema.UserRow.email, "email");
