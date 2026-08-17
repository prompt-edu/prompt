---
paths:
  - "clients/**/*.ts"
  - "clients/**/*.tsx"
---

# React / TypeScript Coding Style

Stack: React 19, TypeScript 6, rspack 2 (Module Federation), Tailwind CSS v4, shadcn/ui + Radix.
Extends `../common/coding-style.md`.

## Naming

- **PascalCase:** React components and component folders (e.g. `ApplicationAssessmentPage`).
- **camelCase:** non-component folders, functions, variables.
- **SCREAMING_SNAKE_CASE:** constants.

## Types

- Prefer `interface` over `type` for object structures.
- **Never use `any`** — strict typing is enforced.
- Place type definitions at the top of the file.

## Imports — shared code comes from the published packages

```typescript
import { Button, ManagementPageHeader } from "@tumaet/prompt-ui-components";
import { useGetCoursePhase, axiosInstance, CoursePhase } from "@tumaet/prompt-shared-state";
```

There is no `@/` alias and no in-repo shared library: everything shared comes from
`@tumaet/prompt-ui-components` or `@tumaet/prompt-shared-state` (see `shared-libraries.md`), and
prefer these over custom UI. Within a component, import its own modules by relative path.

## Folder structure per component

```text
components/   # sub-components
hooks/
interfaces/
pages/        # top level only
utils/
zustand/
```
