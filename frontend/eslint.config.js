import js from "@eslint/js";
import tseslint from "@typescript-eslint/eslint-plugin";
import tsparser from "@typescript-eslint/parser";
import svelte from "eslint-plugin-svelte";
import prettier from "eslint-config-prettier";
import globals from "globals";

/** @type {import('eslint').Linter.Config[]} */
export default [
  // 1. 全局忽略
  {
    ignores: [
      ".svelte-kit/**",
      ".deno-deploy/**",
      "coverage/**",
      "playwright-report/**",
      "test-results/**",
      "node_modules/**",
      "deno-dev.ts",
      "dist/**",
    ],
  },

  // 2. 基础配置 (JS & Svelte)
  js.configs.recommended,
  ...svelte.configs["flat/recommended"],

  // 3. TypeScript 通用配置 (.ts, .js, .svelte.ts)
  {
    files: ["**/*.ts", "**/*.js", "**/*.svelte.ts"],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: "latest",
        sourceType: "module",
      },
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
    plugins: {
      "@typescript-eslint": tseslint,
    },
    rules: {
      ...tseslint.configs.recommended.rules,

      // 🔥 核心修复：关闭 no-undef，交给 TypeScript 编译器处理
      // 这解决了 RequestInit, ServiceWorkerGlobalScope 等类型报错
      "no-undef": "off",

      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
      "@typescript-eslint/explicit-function-return-type": "off",
      "@typescript-eslint/triple-slash-reference": "off",
    },
  },

  // 4. Svelte 5 Runes 支持 (.svelte 和 .svelte.ts)
  // 专门用于补充 $state, $derived 等全局变量
  {
    files: ["**/*.svelte", "**/*.svelte.ts"],
    languageOptions: {
      globals: {
        $state: "readonly",
        $derived: "readonly",
        $effect: "readonly",
        $props: "readonly",
        $bindable: "readonly",
        $inspect: "readonly",
        $host: "readonly",
      },
    },
  },

  // 5. Svelte 组件专用配置 (.svelte)
  {
    files: ["**/*.svelte"],
    languageOptions: {
      parserOptions: {
        parser: tsparser,
        extraFileExtensions: [".svelte"],
      },
    },
    plugins: {
      "@typescript-eslint": tseslint,
    },
    rules: {
      ...tseslint.configs.recommended.rules,
      "no-undef": "off", // Svelte 文件中也关闭 no-undef

      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
      "@typescript-eslint/no-explicit-any": "warn",

      "svelte/no-at-html-tags": "warn",
      "svelte/no-navigation-without-resolve": "off",
      "svelte/require-each-key": "warn",
      // 将 prefer-writable-derived 降级为警告，避免构建失败
      "svelte/prefer-writable-derived": "warn",
    },
  },

  // 6. Service Worker 专用配置
  {
    files: ["src/service-worker.ts"],
    languageOptions: {
      globals: {
        ...globals.serviceworker, // 添加 FetchEvent, ExtendableEvent 等
      },
    },
  },

  // 7. 特殊文件覆盖
  {
    files: ["vite.config.ts", "test-setup.ts"],
    rules: {
      "@typescript-eslint/no-unused-vars": "off",
    },
  },

  // 8. Prettier (最后)
  prettier,
];
