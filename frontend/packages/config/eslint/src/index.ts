import { tanstackConfig } from "@tanstack/eslint-config";

interface WorkspaceEslintConfigOptions {
  tsconfigRootDir: string;
  tsconfigProject?: string | string[];
}

export function createWorkspaceEslintConfig({
  tsconfigRootDir,
  tsconfigProject = ["./tsconfig.app.json", "./tsconfig.node.json"],
}: WorkspaceEslintConfigOptions) {
  return [
    ...tanstackConfig,
    {
      files: ["**/*.{ts,tsx}"],
      languageOptions: {
        parserOptions: {
          project: tsconfigProject,
          tsconfigRootDir,
        },
      },
    },
    {
      rules: {
        semi: ["error", "always"],
        "@typescript-eslint/naming-convention": [
          "error",
          {
            selector: "variable",
            types: ["boolean"],
            format: ["StrictPascalCase"],
            prefix: [
              "is",
              "should",
              "has",
              "can",
              "could",
              "did",
              "will",
              "allow",
              "allows",
              "enable",
              "show",
            ],
          },
          {
            selector: "parameter",
            types: ["boolean"],
            format: ["StrictPascalCase"],
            prefix: [
              "is",
              "should",
              "has",
              "can",
              "could",
              "did",
              "will",
              "allow",
              "allows",
              "enable",
              "show",
            ],
          },
        ],
        "import/no-cycle": "off",
        "import/order": "off",
        "sort-imports": "off",
        "@typescript-eslint/array-type": "off",
        "@typescript-eslint/require-await": "off",
        "@typescript-eslint/consistent-type-imports": ["error", { fixStyle: "separate-type-imports" }],
        "import/consistent-type-specifier-style": ["error", "prefer-top-level"],
        "pnpm/json-enforce-catalog": "off",
      },
    },
    {
      ignores: ["eslint.config.js", "prettier.config.js"],
    },
  ];
}
