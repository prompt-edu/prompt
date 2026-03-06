// @ts-check
// `@type` JSDoc annotations allow editor autocompletion and type checking
// (when paired with `@ts-check`).
// There are various equivalent ways to declare your Docusaurus config.
// See: https://docusaurus.io/docs/api/docusaurus-config

import {themes as prismThemes} from 'prism-react-renderer';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const DISCORD_INVITE_URL = 'https://discord.gg/eybNUqD8gf';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'PROMPT Documentation',
  tagline: 'Project-Oriented Modular Platform for Teaching',
  favicon: 'img/prompt_logo.svg',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  // Set the production url of your site here
  url: 'https://ls1intum.github.io',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/prompt2/',

  // GitHub pages deployment config.
  organizationName: 'ls1intum', // Your GitHub org/user name.
  projectName: 'prompt2', // Your repo name.

  onBrokenLinks: 'throw',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        // We disable the single `docs` preset instance below and register
        // multiple instances of the docs plugin (user, contributor, admin)
        // so each section can live at a top-level route.
        docs: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  // Register separate docs plugin instances for each top-level guide.
  plugins: [
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'user',
        path: 'user',
        routeBasePath: 'user',
        sidebarPath: './sidebars.js',
        editUrl: 'https://github.com/ls1intum/prompt2/tree/main/docs/user/',
      },
    ],
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'contributor',
        path: 'contributor',
        routeBasePath: 'contributor',
        sidebarPath: './sidebars.js',
        editUrl: 'https://github.com/ls1intum/prompt2/tree/main/docs/contributor/',
      },
    ],
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'admin',
        path: 'admin',
        routeBasePath: 'admin',
        sidebarPath: './sidebars.js',
        editUrl: 'https://github.com/ls1intum/prompt2/tree/main/docs/admin/',
      },
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // Replace with your project's social card
      image: 'img/docusaurus-social-card.jpg',
      colorMode: {
        respectPrefersColorScheme: true,
      },
      navbar: {
        title: 'PROMPT',
        logo: {
          alt: 'Prompt Logo',
          src: 'img/prompt_logo.svg',
          // dark-theme variant (falls back to `src` if missing)
          srcDark: 'img/prompt_logo.svg',
          // make the logo a reasonable size in the navbar
          width: 32,
          height: 32,
          // clicking the logo should go to the site root (baseUrl will be applied)
          href: '/',
        },
        items: [
          {
            to: '/user',
            label: 'User Guide',
            position: 'left',
          },
          {
            to: '/contributor',
            label: 'Contributor Guide',
            position: 'left',
          },
          {
            to: '/admin',
            label: 'Administrator Guide',
            position: 'left',
          },
          {
            href: DISCORD_INVITE_URL,
            label: 'Discord',
            position: 'right',
          },
          {
            href: 'https://github.com/ls1intum/prompt2/tree/main',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'User Guide',
                to: '/user',
              },
              {
                label: 'Contributor Guide',
                to: '/contributor',
              },
              {
                label: 'Administrator Guide',
                to: '/admin',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'Discord Community',
                href: DISCORD_INVITE_URL,
              },
              {
                label: 'GitHub Issues',
                href: 'https://github.com/ls1intum/prompt2/issues',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/ls1intum/prompt2/tree/main',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} TUM Applied Education Technologies`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
      },
    }),
};

export default config;
