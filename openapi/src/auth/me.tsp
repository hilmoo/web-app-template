import "@typespec/openapi";
import "@typespec/openapi3";
import "../_generated/db.augmented.tsp";

namespace AuthMe;

@doc("A user account.")
model User is OmitProperties<DbSchema.UserRow, "passwordHash">;

model UpdateUser extends FilterVisibility<
  OptionalProperties<User>,
  #{ all: #[Lifecycle.Update] },
  "UpdateUser"
> {}
