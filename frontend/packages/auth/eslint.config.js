import { defineConfig } from "eslint/config";
import { createWorkspaceEslintConfig } from "@pdv/eslint";

export default defineConfig([
  {
    ignores: ["vitest.config.ts"],
  },
  ...createWorkspaceEslintConfig({
    tsconfigRootDir: import.meta.dirname,
    tsconfigProject: "./tsconfig.json",
  }),
]);
