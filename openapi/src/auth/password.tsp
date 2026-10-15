import "@typespec/openapi";
import "@typespec/openapi3";

import "./me.tsp";

namespace AuthPassword;

model password {
  @doc("The account password")
  @example("hunter2-correct-horse")
  password: string;
}

@doc("Create a new user account with an email and password.")
model RegisterRequest {
  ...FilterVisibility<
    AuthMe.User,
    #{ all: #[Lifecycle.Create] },
    "RegisterRequest"
  >;
  ...password;
}

@doc("Authenticate with an email and password.")
model LoginRequest {
  ...FilterVisibility<
    OmitProperties<AuthMe.User, "fullName">,
    #{ all: #[Lifecycle.Create] },
    "LoginRequest"
  >;
  ...password;
}
