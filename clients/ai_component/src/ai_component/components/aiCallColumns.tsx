import { Badge, Button, type PromptTableColumnDef } from '@tumaet/prompt-ui-components'
import { type AICall, AICallOutcome } from '../interfaces/aiCall'

const tokens = (call: AICall): string =>
  call.promptTokens === null ? '-' : `${call.promptTokens} / ${call.completionTokens ?? '-'}`

export const getAICallColumns = (
  open: (callId: string) => void,
): PromptTableColumnDef<AICall>[] => [
  {
    accessorKey: 'requestedAt',
    header: 'Time',
    cell: ({ row }) => new Date(row.original.requestedAt).toLocaleString(),
  },
  {
    accessorKey: 'feature',
    header: 'Feature',
  },
  {
    accessorKey: 'actorRole',
    header: 'Role',
  },
  {
    id: 'model',
    header: 'Model',
    accessorFn: (call) => call.servedModel ?? call.requestedModel ?? '',
  },
  {
    accessorKey: 'outcome',
    header: 'Outcome',
    cell: ({ row }) => (
      <Badge variant={row.original.outcome === AICallOutcome.SUCCESS ? 'secondary' : 'destructive'}>
        {row.original.outcome}
      </Badge>
    ),
  },
  {
    id: 'tokens',
    header: 'Tokens (prompt / completion)',
    accessorFn: tokens,
  },
  {
    id: 'open',
    header: '',
    cell: ({ row }) => (
      <Button variant='outline' size='sm' onClick={() => open(row.original.id)}>
        Open
      </Button>
    ),
  },
]
