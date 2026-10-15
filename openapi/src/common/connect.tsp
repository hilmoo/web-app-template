import "@typespec/http";

using TypeSpec.Http;

namespace CommonConnect;

@doc("Empty JSON object returned by Connect RPC methods without a payload.")
model Empty {}
