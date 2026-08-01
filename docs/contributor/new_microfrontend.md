---
sidebar_position: 5
---

# How to Add a New Microfrontend

:::info Superseded

This guide has been replaced by **[How to Create a New Course Phase](./new_course_phase.md)**,
which covers the full picture: the micro-frontend, the Go phase service, core registration, and
deployment — plus the `make new-phase` generator that automates the scaffolding.

:::

Key facts that changed since this page was written:

- Build configs are now `rspack.config.mjs` (Rspack), not `webpack.config.mjs`.
- The reference implementation is `clients/example_component` + `servers/example_server`
  (formerly `template_component`/`template_server`); both have READMEs describing their layout.
- External (out-of-repo) course phases **are** supported — see
  [prompt-intro-course](https://github.com/prompt-edu/prompt-intro-course) and
  [prompt-github-challenge](https://github.com/prompt-edu/prompt-github-challenge), and the
  external-phase section of the new guide.
