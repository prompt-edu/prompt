import { defineConfig } from 'vitest/config'

export default defineConfig({
  // @tumaet packages ship sourcemaps whose sources are not published, which Vite warns about per file
  logLevel: 'error',
  // Tests never import CSS, and loading the workspace postcss config warns about its module type
  css: { postcss: {} },
  test: {
    environment: 'node',
    include: ['**/*.test.{ts,tsx}'],
    setupFiles: ['./vitest.setup.ts'],
  },
})
