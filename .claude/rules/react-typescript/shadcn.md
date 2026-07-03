---
paths:
  - "clients/**/*.ts"
  - "clients/**/*.tsx"
---

# shadcn/ui Components

The design system is shadcn/ui + Radix on Tailwind CSS v4, owned centrally — not per component.

- **Install primitives into the shared library only:** `cd clients/shared_library && yarn dlx shadcn add <name>`.
  Components land in `shared_library/components/ui/` and are available to every micro-frontend.
  Never run `shadcn add` inside an individual `*_component` or hand-copy a primitive into one.
- **Import from the published package**, not relative paths:
  `import { Button, Dialog } from '@tumaet/prompt-ui-components'`. Within the shared library itself
  use the `@/components/ui/*` alias.
- **Check before adding.** Most primitives already exist (Button, Card, Dialog, Table, Form, Select,
  Tabs, Tooltip, …) plus custom components (`PromptTable`, `ManagementPageHeader`, `MultiSelect`,
  `DatePicker`, `DeleteConfirmation`). See `shared-libraries.md`. Reuse before installing/building.
- **Styling:** compose classes with `cn` (clsx + tailwind-merge); use Tailwind v4 tokens, not
  hard-coded colors. Keep Radix accessibility props intact; don't strip `aria-*`/`role`.
- For the full workflow use the `add-shared-ui-component` skill.
