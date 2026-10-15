import "@typespec/http";
import "@typespec/openapi";
import "@typespec/openapi3";

import "../common/error.tsp";
import "../common/connect.tsp";
import "./password.tsp";
import "./me.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace CoreServices;

@tag("Auth")
interface Auth {
  @doc("Authenticate a user with email and password, returning a session cookie on success.")
  @post
  @route("/services.v1.AuthService/AuthLogin")
  login(@body body: AuthPassword.LoginRequest):
    | AuthMe.User
    | CommonError.CreateCommonErrorAuth;

  @doc("Create a new user account with an email and password, returning a session cookie on success.")
  @post
  @route("/services.v1.AuthService/AuthRegister")
  register(@body body: AuthPassword.RegisterRequest):
    | AuthMe.User
    | CommonError.CreateCommonErrorAuth;

  @doc("Get the currently authenticated user's profile.")
  @post
  @route("/services.v1.AuthService/AuthMe")
  me(): AuthMe.User | CommonError.GetCommonErrorAuth;

  @doc("Update the currently authenticated user's profile.")
  @post
  @route("/services.v1.AuthService/AuthUpdateMe")
  updateMe(@body body: AuthMe.UpdateUser):
    | AuthMe.User
    | CommonError.UpdateCommonErrorAuth;

  @doc("Log out the currently authenticated user, invalidating the session cookie.")
  @post
  @route("/services.v1.AuthService/AuthLogout")
  logout(): CommonConnect.Empty | CommonError.CreateCommonErrorAuth;
}
