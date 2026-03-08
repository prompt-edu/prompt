import { fixupConfigRules, fixupPluginRules } from '@eslint/compat'
import typescriptEslint from '@typescript-eslint/eslint-plugin'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import globals from 'globals'
import tsParser from '@typescript-eslint/parser'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import js from '@eslint/js'
import { FlatCompat } from '@eslint/eslintrc'
import fs from 'fs'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
})

const getTsConfigPaths = () => {
  const workspaceFolders = [
    'core',
    'template_component',
    'shared_library',
    'interview_component',
    'matching_component',
    'intro_course_tutor_component',
    'intro_course_developer_component',
    'assessment_component',
    'devops_challenge_component',
    'team_allocation_component',
    'self_team_allocation_component',
    'certificate_component',
  ] // TODO: replace with dynamic workspace detection
  return workspaceFolders
    .map((folder) => {
      const configPath = path.resolve(__dirname, folder, 'tsconfig.json')
      return fs.existsSync(configPath) ? configPath : null
    })
    .filter(Boolean)
}

export default [
  ...fixupConfigRules(
    compat.extends(
      'plugin:react/recommended',
      'plugin:@typescript-eslint/recommended',
      'prettier',
      'plugin:prettier/recommended',
    ),
  ),
  {
    files: ['**/*.js', '**/*.jsx', '**/*.ts', '**/*.tsx'],
    ignores: [
      'node_modules/**', // Ignore node_modules
      '**/dist/**', // Ignore distribution folders
      '**/build/**', // Ignore build folders
      '**/postcss.config.js',
      '**/env.js',
      '**/env.template.js',
    ],

    plugins: {
      '@typescript-eslint': fixupPluginRules(typescriptEslint),
      react: fixupPluginRules(react),
      'react-hooks': fixupPluginRules(reactHooks),
    },

    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.jest,
      },

      parser: tsParser,
      ecmaVersion: 'latest',
      sourceType: 'module',

      parserOptions: {
        ecmaFeatures: {
          jsx: true,
        },

        project: getTsConfigPaths(), // Dynamically resolve all tsconfig.json files
      },
    },

    settings: {
      react: {
        version: 'detect',
      },
    },

    rules: {
      'no-shadow': 'off',
      '@typescript-eslint/no-shadow': ['error'],

      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/strict-boolean-expressions': 'off',
      '@typescript-eslint/return-await': 'off',

      'max-len': [
        'warn',
        {
          code: 160,
          ignoreComments: true,
          ignoreUrls: true,
        },
      ],

      'react/react-in-jsx-scope': 'off',

      'react/jsx-filename-extension': [
        1,
        {
          extensions: ['.tsx', '.jsx'],
        },
      ],

      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',

      'prettier/prettier': [
        'error',
        {
          endOfLine: 'auto',
        },
      ],
    },
  },
]
