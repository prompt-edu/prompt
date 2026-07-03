---
paths:
  - "clients/**/*.ts"
  - "clients/**/*.tsx"
---

# Client Shared Libraries — Prefer These Over Custom Code

Always check and reuse these before writing custom components/functions.

## `@tumaet/prompt-ui-components` (source: `../prompt-lib/prompt-ui-components/`)

```typescript
import { Button, Card, ManagementPageHeader } from '@tumaet/prompt-ui-components'
```

- **shadcn/ui primitives:** Accordion, Alert, AlertDialog, Avatar, Badge, Breadcrumb, Button,
  Calendar, Card (+ Header/Title/Description/Content/Footer), Chart, Checkbox, Collapsible, Command,
  Dialog, DropdownMenu, Form, Input, Label, Popover, Progress, RadioGroup, ScrollArea, Select,
  Separator, Sheet, Sidebar, Skeleton, Switch, Table, Tabs, Textarea, Toast, Toaster, Toggle,
  ToggleGroup, Tooltip.
- **Pages:** `ManagementPageHeader`, `ErrorPage`, `LoadingPage`, `UnauthorizedPage`, `SaveChangesAlert`.
- **Dialogs:** `DeleteConfirmation`, `DialogErrorDisplay`, `DialogLoadingDisplay`.
- **Inputs:** `DatePicker`, `DatePickerWithRange`, `MultiSelect`.
- **Data table:** `PromptTable<T>` — sorting, filtering, row selection, column visibility, search, row actions.
- **Rich text:** `MinimalTiptapEditor`, `MailingTiptapEditor`, `DescriptionMinimalTiptapEditor`.
- **Hooks:** `useToast`, `useIsMobile`, `useCustomElementWidth`, `useScreenSize`.
- **Utilities:** `cn` (clsx + tailwind-merge). **Table types:** `WithId`, `RowAction<T>`,
  `TableFilter`, `SortableHeader`.

## `@tumaet/prompt-shared-state` (source: `prompt-shared-state` repo)

```typescript
import { Role, PassStatus, useCourseStore } from '@tumaet/prompt-shared-state'
```

- **Enums:** `Role`, `PassStatus`, `ScoreLevel`, `CourseType`, `Gender`, `StudyDegree`.
- **Domain types:** `Course`, `CoursePhaseWithMetaData`, `CoursePhaseWithType`, `CreateCoursePhase`,
  `UpdateCoursePhase`, `CoursePhaseParticipationWithStudent`, `Student`, `Person`, `Team`, `User`, …
- **Utils:** `mapScoreLevelToNumber`, `mapNumberToScoreLevel`, `getPermissionString`,
  `getGenderString`, `getStudyDegreeString`.
- **Zustand stores:** `useAuthStore` (`user`, `permissions`, `logout`, `setPermissions`),
  `useCourseStore` (`courses`, `ownCourseIDs`, `setSelectedCourseID`, `isStudentOfCourse`, …).

## In-repo `shared_library` (`@/` alias)

- **Hooks:** `useGetCoursePhase`, `useModifyCoursePhase`, `useUpdateCoursePhaseParticipation`,
  `useUpdateCoursePhaseParticipationBatch`, `useUpdateCoursePhaseMetaData`, `useGetMailingIsConfigured`.
- **Queries:** `getCoursePhase`, `getCoursePhaseParticipations`, `getOwnCoursePhaseParticipation`,
  `getCoursePhaseParticipationStatusCounts`.
- **Mutations:** `updateCoursePhase` (JSON-Patch), `updateCoursePhaseParticipation(Batch)`, `sendStatusMail`.
- **Page components:** `CoursePhaseParticipationsTable`, `CoursePhaseMailing`.
- **UI:** `StudentProfile`, `StudentAvatar`/`RenderStudents`, `SettingsCard`, `FilterBadge`,
  `DynamicIcon`, `UnauthorizedPage`, `MissingConfig`/`MissingSettings`, `ExportedApplicationAnswerTable`,
  `SortableHeader`, `GroupActionDialog`.
- **Utils:** `getStatusBadge`/`getStatusString`/`getStatusColor`, `getGravatarUrl`,
  `getCountryName`/`countriesArr`, `axiosInstance`, `env`.

## Adding components

`cd clients/shared_library && yarn dlx shadcn add <component>` — see the `add-shared-ui-component` skill.
