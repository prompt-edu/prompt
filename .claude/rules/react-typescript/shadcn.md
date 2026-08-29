---
paths:
  - "clients/**/*.ts"
  - "clients/**/*.tsx"
---

# shadcn/ui Components

The design system is shadcn/ui + Radix on Tailwind CSS v4, owned centrally by the
`prompt-edu/prompt-lib` repository and consumed here as `@tumaet/prompt-ui-components`. It is not
owned per component, and this repository holds no primitives of its own.

- **Never run `shadcn add` inside a `*_component`** and never hand-copy a primitive into one. New
  primitives are added in `prompt-edu/prompt-lib` and arrive here as a published package version.
  `clients/core/components.json` exists only so the shadcn CLI resolves every alias to
  `@tumaet/prompt-ui-components`.
- **Import from the published package**, not relative paths:
  `import { Button, Dialog } from '@tumaet/prompt-ui-components'`.
- **Check before adding.** Most primitives already exist (Button, Card, Dialog, Table, Form, Select,
  Tabs, Tooltip, …) plus custom components (`PromptTable`, `ManagementPageHeader`, `MultiSelect`,
  `DatePicker`, `DeleteConfirmation`). See `shared-libraries.md`. Reuse before installing/building.
- **Styling:** compose classes with `cn` (clsx + tailwind-merge); use Tailwind v4 tokens, not
  hard-coded colors. Keep Radix accessibility props intact; don't strip `aria-*`/`role`.
- **One Tailwind build.** `clients/core` compiles the only stylesheet and scans every component
  directory; a `*_component` never declares a Tailwind entry point. See the styling section of
  `../module-federation/remotes.md`.
- For the full workflow use the `add-shared-ui-component` skill.
