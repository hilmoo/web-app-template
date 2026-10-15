import type { OpenApiDocument } from "orval";

const camel = (s: string) => s.replace(/_+([a-z0-9])/gi, (_, c: string) => c.toUpperCase());

// Recursively rename object property names (and `required` entries) to camelCase so the
// generated schemas match the camelCase fields used by the ConnectRPC messages.
function walk(node: unknown): void {
  if (Array.isArray(node)) {
    return node.forEach(walk);
  }
  if (!node || typeof node !== "object") {
    return;
  }
  const obj = node as Record<string, unknown>;
  const props = obj.properties;
  if (props && typeof props === "object" && !Array.isArray(props)) {
    obj.properties = Object.fromEntries(Object.entries(props).map(([k, v]) => [camel(k), v]));
  }
  if (Array.isArray(obj.required)) {
    obj.required = obj.required.map((k) => (typeof k === "string" ? camel(k) : k));
  }
  Object.values(obj).forEach(walk);
}

export default function transformer(spec: OpenApiDocument): OpenApiDocument {
  // Only touch schemas; parameter names (query/path/header) must stay as the server expects.
  walk(spec.components?.schemas);
  return spec;
}
