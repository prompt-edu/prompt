---
name: add-shared-ui-component
description: Add or reuse a shared UI component in PROMPT 2.0 — install shadcn/ui primitives into the shared_library and surface them via @tumaet/prompt-ui-components instead of writing custom UI. Use when adding a button/dialog/table/form or any reusable UI element.
---

PROMPT 2.0 has a large shared UI library. ALWAYS prefer reusing it over writing custom components.

## 1. Check what already exists first

Before adding anything, search the catalog in `.claude/rules/react-typescript/shared-libraries.md`
and `AGENTS.md`. The library already ships most shadcn/ui primitives (Button, Card, Dialog, Table,
Form, Select, Tabs, Tooltip, …) plus custom components (`ManagementPageHeader`, `PromptTable<T>`,
`DatePicker`, `MultiSelect`, rich-text editors, `DeleteConfirmation`, …). If it exists, import it:

```typescript
import { Button, PromptTable, ManagementPageHeader } from '@tumaet/prompt-ui-components'
```

Project-internal shared code uses the `@/` alias (`@/components/ui/*`, `@/hooks/*`, `@/network/*`).

## 2. Add a missing shadcn/ui primitive

```bash
cd clients/shared_library
yarn dlx shadcn add <component-name>
```

Components land in `shared_library/components/ui/` and become available to every micro-frontend.

## 3. Surface a custom shared component

If you build a genuinely new reusable component, add it to the shared library
(`../prompt-lib/prompt-ui-components/`, published as `@tumaet/prompt-ui-components`) so all phases
can use it — do not duplicate it inside a single component.

## Verify

- Import resolves from `@tumaet/prompt-ui-components` (or `@/...` for in-repo shared code).
- `make lint` passes; no `any`; styling via `cn` (clsx + tailwind-merge), Tailwind v4 tokens.
