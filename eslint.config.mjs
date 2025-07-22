import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";

export default [
  {
    ignores: [
      "app/assets/builds/**/*",
      "tailwind.config.js",
      "coverage/**/*",
      "public/**/*",
      "vendor/**/*",
      "tmp/**/*",
      "log/**/*",
      "node_modules/**/*",
    ],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["**/*.{js,mjs,cjs,ts,mts,cts}"],
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
        $: "readonly",
      },
    },
  },
];
