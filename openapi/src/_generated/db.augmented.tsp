// Hand-written layer on top of the generated db.tsp: visibility rules and formats
// that cannot be expressed in the drizzle schema.
import "./db.tsp";

namespace DbSchema;

@@format(DbSchema.UserRow.email, "email");
