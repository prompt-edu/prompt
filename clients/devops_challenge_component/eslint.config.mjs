import rootConfig from '../eslint.config.mjs'

export default [
  ...rootConfig,
  {
    // Optionally add any subfolder-specific rules or settings
    files: ['**/*.ts', '**/*.tsx'],
    rules: {},
  },
]
