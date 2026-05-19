import sharedConfig from '@tumaet/prompt-ui-components/tailwind-config'

/** @type {import('tailwindcss').Config} */
export default {
  presets: [sharedConfig],
  content: [
    './src/**/*.{js,ts,jsx,tsx}',
    '../node_modules/@tumaet/prompt-ui-components/dist/**/*.{js,ts,jsx,tsx}',
  ],
}
