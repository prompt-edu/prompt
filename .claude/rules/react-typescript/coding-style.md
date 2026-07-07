---
paths:
  - "clients/**/*.ts"
  - "clients/**/*.tsx"
---

# React / TypeScript Coding Style

Stack: React 19, TypeScript 5.9, Webpack 5 (Module Federation), Tailwind CSS v4, shadcn/ui + Radix.
Extends `../common/coding-style.md`.

## Naming

- **PascalCase:** React components and component folders (e.g. `ApplicationAssessmentPage`).
- **camelCase:** non-component folders, functions, variables.
- **SCREAMING_SNAKE_CASE:** constants.

## Types

- Prefer `interface` over `type` for object structures.
- **Never use `any`** — strict typing is enforced.
- Place type definitions at the top of the file.

## Imports — use the `@/` alias for in-repo shared code

```typescript
import { Button } from "@/components/ui/button";        // shadcn/ui
import { useGetCoursePhase } from "@/hooks/useGetCoursePhase";
import { getCoursePhase } from "@/network/queries/getCoursePhase";
import { CoursePhase } from "@/interfaces/coursePhase";
```

Shared, published code comes from `@tumaet/prompt-ui-components` and `@tumaet/prompt-shared-state`
(see `shared-libraries.md`). Prefer these over custom UI.

## Folder structure per component

```text
components/   # sub-components
hooks/
interfaces/
pages/        # top level only
utils/
zustand/
```
