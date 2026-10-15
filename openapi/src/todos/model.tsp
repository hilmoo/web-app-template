import "@typespec/openapi";
import "@typespec/openapi3";
import "../_generated/db.augmented.tsp";

namespace TodoModel;

model Id {
  @doc("The ID of the todo.")
  @example("123e4567-e89b-12d3-a456-426614174000")
  todoId: string;
}

@doc("A user todo.")
model Base is DbSchema.TodoRow;

@doc("A request to create a new todo.")
model CreateRequest is FilterVisibility<
  OmitProperties<Base, "ownerId" | "done">,
  #{ all: #[Lifecycle.Create] },
  "CreateRequest"
>;

@doc("A request to update a todo.")
model UpdateRequest {
  ...Id;
  ...FilterVisibility<
    OptionalProperties<OmitProperties<Base, "ownerId">>,
    #{ all: #[Lifecycle.Update] },
    "UpdateRequest"
  >;
}
