import { defineConfig } from "eslint/config";
import { createWorkspaceEslintConfig } from "@pdv/eslint";

export default defineConfig([
  ...createWorkspaceEslintConfig({
    tsconfigRootDir: import.meta.dirname,
    tsconfigProject: "./tsconfig.json",
  }),
  {
    rules: {
      "@typescript-eslint/naming-convention": "off",
    },
  },
]);
