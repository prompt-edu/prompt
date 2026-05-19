import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Button,
  Card,
  CardContent,
  DeleteConfirmation,
  useToast,
} from '@tumaet/prompt-ui-components'
import { ChevronDown, ChevronRight, ExternalLink, RotateCcw, Trash2 } from 'lucide-react'

import { ResourceInstance } from '../interfaces/resourceInstance'
import { retryInstance } from '../network/mutations/retryInstance'
import { deleteInstance } from '../network/mutations/deleteInstance'
import { StatusBadge } from './StatusBadge'

interface Props {
  coursePhaseID: string
  instance: ResourceInstance
}

export const InstanceRow = ({ coursePhaseID, instance }: Props) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [expanded, setExpanded] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)

  const onMutationError = (action: string) => (err: unknown) => {
    toast({
      title: `Failed to ${action} instance`,
      description: err instanceof Error ? err.message : 'Unknown error',
      variant: 'destructive',
    })
  }

  const { mutate: retry, isPending: isRetrying } = useMutation({
    mutationFn: () => retryInstance(coursePhaseID, instance.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances', coursePhaseID] })
      toast({ title: 'Retry started' })
    },
    onError: onMutationError('retry'),
  })

  const { mutate: remove, isPending: isDeleting } = useMutation({
    mutationFn: () => deleteInstance(coursePhaseID, instance.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances', coursePhaseID] })
      toast({ title: 'Instance deleted' })
      setConfirmOpen(false)
    },
    onError: onMutationError('delete'),
  })

  const hasError = !!instance.errorMessage && instance.errorMessage.length > 0
  const target = instance.teamId
    ? `team ${instance.teamId.slice(0, 8)}…`
    : instance.courseParticipationId
      ? `student ${instance.courseParticipationId.slice(0, 8)}…`
      : 'unassigned'

  return (
    <>
      <Card>
        <CardContent className='space-y-2 p-4'>
          <div className='flex items-start justify-between gap-4'>
            <div className='space-y-1'>
              <div className='flex flex-wrap items-center gap-2'>
                <StatusBadge status={instance.status} />
                <span className='font-mono text-xs text-muted-foreground'>{instance.id}</span>
              </div>
              <div className='text-sm'>
                target: <span className='font-mono'>{target}</span>
                {instance.retryCount > 0 && (
                  <span className='ml-2 text-xs text-muted-foreground'>
                    retries: {instance.retryCount}
                  </span>
                )}
              </div>
              {instance.externalUrl && (
                <a
                  href={instance.externalUrl}
                  target='_blank'
                  rel='noopener noreferrer'
                  className='inline-flex items-center gap-1 text-sm text-blue-600 hover:underline'
                >
                  <ExternalLink className='h-3 w-3' />
                  {instance.externalUrl}
                </a>
              )}
            </div>

            <div className='flex items-center gap-2'>
              {instance.status === 'failed' && (
                <Button variant='outline' size='sm' onClick={() => retry()} disabled={isRetrying}>
                  <RotateCcw className='mr-1 h-3 w-3' />
                  {isRetrying ? 'Retrying…' : 'Retry'}
                </Button>
              )}
              <Button
                variant='outline'
                size='sm'
                onClick={() => setConfirmOpen(true)}
                disabled={isDeleting}
              >
                <Trash2 className='mr-1 h-3 w-3' /> Delete
              </Button>
            </div>
          </div>

          {hasError && (
            <div>
              <button
                type='button'
                onClick={() => setExpanded((p) => !p)}
                className='inline-flex items-center gap-1 text-sm text-red-700 hover:underline'
              >
                {expanded ? (
                  <ChevronDown className='h-3 w-3' />
                ) : (
                  <ChevronRight className='h-3 w-3' />
                )}
                Error details
              </button>
              {expanded && (
                <pre className='mt-2 overflow-x-auto rounded bg-red-50 p-2 text-xs text-red-900'>
                  {instance.errorMessage}
                </pre>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {confirmOpen && (
        <DeleteConfirmation
          isOpen={confirmOpen}
          setOpen={setConfirmOpen}
          deleteMessage='Delete this resource instance?'
          customWarning='This only removes the row in PROMPT. The external resource (if any) is NOT deleted.'
          onClick={(confirmed) => {
            if (confirmed) remove()
          }}
        />
      )}
    </>
  )
}

export default InstanceRow
