import { readdirSync } from "node:fs";
import resolve from "@rollup/plugin-node-resolve";
import commonjs from "@rollup/plugin-commonjs";
import dts from "rollup-plugin-dts";

// index + one entry per generated namespace barrel
const namespaces = readdirSync("./namespaces")
  .filter((f) => f.endsWith(".js"))
  .map((f) => f.slice(0, -3));

const jsInputs = {
  index: "./index.js",
  ...Object.fromEntries(namespaces.map((n) => [`namespaces/${n}`, `./namespaces/${n}.js`])),
};

const dtsInputs = {
  index: "./index.d.ts",
  ...Object.fromEntries(namespaces.map((n) => [`namespaces/${n}`, `./namespaces/${n}.d.ts`])),
};

export default [
  {
    input: jsInputs,
    output: {
      dir: "dist",
      format: "esm",
      sourcemap: true,
      entryFileNames: "[name].mjs",
      chunkFileNames: "chunks/[name]-[hash].mjs",
    },
    plugins: [resolve(), commonjs()],
    external: ["@bufbuild/protobuf"],
  },
  {
    input: dtsInputs,
    output: {
      dir: "dist",
      format: "es",
      entryFileNames: "[name].d.ts",
      chunkFileNames: "chunks/[name]-[hash].d.ts",
    },
    plugins: [dts()],
  },
];
