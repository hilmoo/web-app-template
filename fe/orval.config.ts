import { defineConfig } from "orval";

export default defineConfig({
  api: {
    input: {
      target: "../openapi/openapi.json",
      override: { transformer: "./orval.transformer.ts" },
    },
    output: {
      client: "zod",
      mode: "tags-split",
      target: "./src/api/schemas",
      override: {
        zod: {
          version: 4,
        },
      },
    },
  },
});
