import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitest/config'

const coreSrc = fileURLToPath(new URL('./core/src', import.meta.url))

export default defineConfig({
  // @tumaet packages ship sourcemaps whose sources are not published, which Vite warns about per file
  logLevel: 'error',
  // Tests never import CSS, and loading the workspace postcss config warns about its module type
  css: { postcss: {} },
  // The path aliases core declares in its tsconfig. rspack resolves them for the build; without
  // them here a test that reaches a module importing one by value cannot load it.
  resolve: {
    alias: [
      { find: /^@core\//, replacement: `${coreSrc}/` },
      { find: /^@managementConsole\//, replacement: `${coreSrc}/managementConsole/` },
    ],
  },
  test: {
    environment: 'node',
    include: ['**/*.test.{ts,tsx}'],
    setupFiles: ['./vitest.setup.ts'],
  },
})
