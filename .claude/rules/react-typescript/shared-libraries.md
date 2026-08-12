---
paths:
  - "clients/**/*.ts"
  - "clients/**/*.tsx"
---

# Client Shared Libraries — Prefer These Over Custom Code

Always check and reuse these before writing custom components/functions. Both libraries are external
packages; there is no in-repo shared library and no `@/` alias — everything shared lives in one of
the two packages below.

## `@tumaet/prompt-ui-components` (source: `github.com/prompt-edu/prompt-lib`)

```typescript
import { Button, Card, ManagementPageHeader } from '@tumaet/prompt-ui-components'
```

- **shadcn/ui primitives:** Accordion, Alert, AlertDialog, Avatar, Badge, Breadcrumb, Button,
  Calendar, Card (+ Header/Title/Description/Content/Footer), Chart, Checkbox, Collapsible, Command,
  Dialog, DropdownMenu, Form, Input, Label, Popover, Progress, RadioGroup, ScrollArea, Select,
  Separator, Sheet, Sidebar, Skeleton, Switch, Table, Tabs, Textarea, Toast, Toaster, Toggle,
  ToggleGroup, Tooltip.
- **Pages:** `ManagementPageHeader`, `ErrorPage`, `LoadingPage`, `UnauthorizedPage`, `SaveChangesAlert`.
- **Dialogs:** `DeleteConfirmation`, `DialogErrorDisplay`, `DialogLoadingDisplay`, `GroupActionDialog`.
- **Inputs:** `DatePicker`, `DatePickerWithRange`, `MultiSelect`, `ScoreLevelSelector`, `FileUpload`,
  `FileList`.
- **Data table:** `PromptTable<T>` — sorting, filtering, row selection, column visibility, search,
  row actions. `PromptTableURL<T>` mirrors it with the table state kept in the URL.
- **Page components:** `CoursePhaseParticipationsTable`, `CoursePhaseMailing`.
- **Student/course UI:** `StudentProfile`, `StudentAvatar`, `StudentProfilePicture`, `SettingsCard`,
  `FilterBadge`, `DynamicIcon`, `MissingConfig`, `MissingSettings`, `ExportedApplicationAnswerTable`,
  `ThemeToggle`.
- **Rich text:** `MinimalTiptapEditor`, `MailingTiptapEditor`, `DescriptionMinimalTiptapEditor`.
- **Hooks:** `useToast`, `useIsMobile`, `useCustomElementWidth`, `useScreenSize`.
- **Utilities:** `cn` (clsx + tailwind-merge), `getStatusBadge`, `getStatusColor`, `getGravatarUrl`,
  `getCountries`, `formatFileSize`, `openFileDownload`. **Table types:** `WithId`, `RowAction<T>`,
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
- **Hooks:** `useGetCoursePhase`, `useGetCoursePhaseParticipants`, `useModifyCoursePhase`,
  `useUpdateCoursePhaseParticipation`, `useUpdateCoursePhaseParticipationBatch`,
  `useUpdateCoursePhaseMetaData`, `useGetMailingIsConfigured`, `useFileUpload`.
- **Queries:** `getCoursePhase`, `getCoursePhaseParticipations`, `getOwnCoursePhaseParticipation`,
  `getCoursePhaseParticipationStatusCounts`.
- **Mutations:** `updateCoursePhase` (JSON-Patch), `updateCoursePhaseParticipationBatch`,
  `updateCoursePhaseParticipationMetaData`, `sendStatusMail`, `uploadFile`, `deleteApplicationFile`.
- **Network/config:** `axiosInstance` (JWT injection + CORS), `configService`, the global `env` object.

## Adding components

New shared UI belongs in the `prompt-edu/prompt-lib` repository and reaches this repo as a published
`@tumaet/prompt-ui-components` version — see the `add-shared-ui-component` skill.
