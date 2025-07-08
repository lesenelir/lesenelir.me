import { base, next } from '@lesenelir/eslint-config'

/**
 * ESLint configuration for the monorepo
 * @type {import("eslint").Linter.Config[]}
 */
export default [
  // Global ignores
  {
    ignores: [
      '**/node_modules/**',
      '**/dist/**',
      '**/build/**',
      '**/.next/**',
      '**/coverage/**',
      '**/public/**',
      '**/*.config.js',
      '**/*.config.mjs',
      '**/*.config.ts',
      'pnpm-lock.yaml'
    ]
  },

  // Base configuration for all TypeScript/JavaScript files
  ...base,

  // Next.js specific configuration for web app
  {
    files: ['apps/web/**/*.{js,jsx,ts,tsx}'],
    ...next[0]
  },

  // Package-specific configurations
  {
    files: ['packages/**/*.{js,jsx,ts,tsx}'],
    rules: {
      // Allow console.log in development packages
      'no-console': 'off'
    }
  }
]
