---
name: add-shared-ui-component
description: Add or reuse a shared UI component in PROMPT 2.0 — reuse @tumaet/prompt-ui-components instead of writing custom UI, and add genuinely new primitives in the prompt-lib repository. Use when adding a button/dialog/table/form or any reusable UI element.
---

PROMPT 2.0 has a large shared UI library. ALWAYS prefer reusing it over writing custom components.

The library is **not** in this repository. It lives in `prompt-edu/prompt-lib` and is consumed here
as the published package `@tumaet/prompt-ui-components`. There is no in-repo shared library and no
`@/` alias.

## 1. Check what already exists first

Before adding anything, search the catalog in `.claude/rules/react-typescript/shared-libraries.md`.
The library already ships most shadcn/ui primitives (Button, Card, Dialog, Table, Form, Select,
Tabs, Tooltip, …) plus custom components (`ManagementPageHeader`, `PromptTable<T>`, `DatePicker`,
`MultiSelect`, rich-text editors, `DeleteConfirmation`, `CoursePhaseParticipationsTable`,
`CoursePhaseMailing`, …). If it exists, import it:

```typescript
import { Button, PromptTable, ManagementPageHeader } from '@tumaet/prompt-ui-components'
```

Shared state, domain types, hooks, queries, and `axiosInstance` come from
`@tumaet/prompt-shared-state`.

## 2. Add a missing shadcn/ui primitive

Never run `shadcn add` inside a `*_component`, and never hand-copy a primitive into one. Add the
primitive in `prompt-edu/prompt-lib`, release it, and bump `@tumaet/prompt-ui-components` in
`clients/package.json`. `clients/core/components.json` only maps every shadcn alias onto the
published package; it is not a place to install components.

## 3. Surface a custom shared component

If you build a genuinely new reusable component, add it to `prompt-edu/prompt-lib` so all phases can
use it — do not duplicate it inside a single component. Something used by exactly one phase belongs
in that phase's own `src/`, not in the shared library.

## Verify

- Import resolves from `@tumaet/prompt-ui-components` or `@tumaet/prompt-shared-state`.
- `make lint` passes; no `any`; styling via `cn` (clsx + tailwind-merge), Tailwind v4 tokens.
