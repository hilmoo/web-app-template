import "@typespec/http";
import "@typespec/openapi";
import "@typespec/openapi3";

import "./auth";
import "./todos";

using TypeSpec.Http;
using TypeSpec.OpenAPI;

@service(#{ title: "Core Service" })
@info(#{ version: "1.0.0" })
@doc("web-app-template")
@server("http://localhost:8080/", "Default server")
namespace CoreServices;
