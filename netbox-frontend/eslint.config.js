/**
 * ESLint flat config for NetBox Go frontend
 *
 * Stack: Vue 3.5 (Composition API + <script setup>) + TypeScript 6 + Vite 8
 * Format: Prettier handles all formatting; ESLint focuses on code quality & bugs.
 *
 * @see https://eslint.org/docs/latest/use/configure/configuration-files
 */
import js from '@eslint/js'
import vue from 'eslint-plugin-vue'
import tseslint from 'typescript-eslint'
import eslintConfigPrettier from 'eslint-config-prettier'
import globals from 'globals'

export default [
  // ── Global ignores ────────────────────────────────────────────────
  {
    ignores: [
      'dist/**',
      'node_modules/**',
      'coverage/**',
      '*.config.{js,ts,cjs,mjs}',
      '.vscode/**',
    ],
  },

  // ── Base JS recommended rules ─────────────────────────────────────
  js.configs.recommended,

  // ── TypeScript recommended (type-aware disabled for build speed) ──
  ...tseslint.configs.recommended,

  // ── Vue: flat/recommended includes both essential + strongly-recommended ──
  ...vue.configs['flat/recommended'],

  // ── Disable ESLint rules that conflict with Prettier formatting ───
  eslintConfigPrettier,

  // ── Project-wide language options ─────────────────────────────────
  {
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.es2021,
        ...globals.node,
      },
    },
  },

  // ── Vue SFC support: parse <template> as HTML, <script> as TS ─────
  {
    files: ['**/*.vue'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
        // Tell the Vue parser where the TS config lives so it can resolve
        // path aliases (@/*) and global type declarations.
        projectService: false,
        extraFileExtensions: ['.vue'],
      },
    },
    rules: {
      // Allow multi-word component names to be optional — our registry
      // intentionally registers some single-word view components.
      'vue/multi-word-component-names': 'off',
      // Enforce a consistent order for <script setup> top-level members.
      'vue/component-api-style': ['error', ['script-setup']],
      // Require a name attribute only on recursive/trees components.
      'vue/require-name-property': 'off',
      // Prettier handles attribute & template indentation; don't double-report.
      'vue/html-indent': 'off',
      'vue/script-indent': 'off',
      'vue/max-attributes-per-line': 'off',
      'vue/singleline-html-element-content-newline': 'off',
      'vue/html-self-closing': 'off',
      // Incompatible with <script setup> + TS — defineProps handles defaults
      // via `withDefaults()`. This rule expects the old options-API style.
      'vue/require-default-prop': 'off',
    },
  },

  // ── TypeScript files ──────────────────────────────────────────────
  {
    files: ['**/*.ts', '**/*.tsx'],
    rules: {
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_',
        },
      ],
      '@typescript-eslint/no-explicit-any': 'error',
    },
  },

  // ── Test files: relax some rules ──────────────────────────────────
  {
    files: ['**/*.test.ts', '**/*.spec.ts'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-non-null-assertion': 'off',
      'no-console': 'off',
    },
  },

  // ── Universal overrides ───────────────────────────────────────────
  {
    rules: {
      // Allow console.* in development; CI can tighten this if desired.
      'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      // Prefer const for refs/computeds that aren't reassigned.
      'prefer-const': 'error',
      // Catch accidental use of == vs ===
      eqeqeq: ['error', 'always', { null: 'ignore' }],
      // No unreachable code (belt-and-suspenders with TS).
      'no-unreachable': 'error',
      // Vue: enforce component naming style (PascalCase) in templates.
      'vue/component-name-in-template-casing': ['error', 'PascalCase'],
      '@typescript-eslint/no-explicit-any': 'error',
      // Allow unused vars prefixed with underscore.
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_',
        },
      ],
      // markdown.ts uses control chars legitimately in regex.
      'no-control-regex': 'off',
    },
  },
]
