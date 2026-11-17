import type { PromptTableColumnDef } from '@tumaet/prompt-ui-components'
import type { OrgRow } from '../utils/orgHierarchy'

export const orgTableColumns: PromptTableColumnDef<OrgRow>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
  },
  {
    accessorKey: 'slug',
    header: 'Slug',
  },
  {
    accessorKey: 'parentLabel',
    header: 'Parent Org',
    cell: ({ row }) => row.original.parentLabel ?? '—',
  },
  {
    accessorKey: 'university',
    header: 'University',
    cell: ({ row }) => row.original.university ?? '—',
  },
]
