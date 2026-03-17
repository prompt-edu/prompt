import sharedConfig from '@tumaet/prompt-ui-components/tailwind-config'

/** @type {import('tailwindcss').Config} */
export const presets = [sharedConfig]
export const content = [
  'src/**/*.{ts,tsx}',
  '../node_modules/@tumaet/prompt-ui-components/dist/**/*.{ts,tsx}',
]
