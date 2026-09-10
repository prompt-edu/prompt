import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Badge,
  Button,
  Card,
  CardContent,
  cn,
  DeleteConfirmation,
  useToast,
} from '@tumaet/prompt-ui-components'
import { ChevronDown, ChevronRight, ExternalLink, RotateCcw, Trash2 } from 'lucide-react'
import { useState } from 'react'

import type { ResourceInstance } from '../interfaces/resourceInstance'
import { deleteInstance } from '../network/mutations/deleteInstance'
import { retryInstance } from '../network/mutations/retryInstance'
import { describeError } from '../utils/describeError'
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
      description: describeError(err),
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
  const isPartial = instance.status === 'partial'
  const isRetryable = instance.status === 'failed' || isPartial
  const targetKind = instance.scope === 'per_team' ? 'team' : 'student'
  const targetID = instance.teamId ?? instance.courseParticipationId
  const target = instance.targetName || (targetID ? `${targetKind} ${targetID}` : 'unassigned')
  // Both names are only filled once the worker has run the instance.
  const resourceName = instance.resolvedName || instance.nameTemplate

  return (
    <>
      <Card>
        <CardContent className='space-y-2 p-4'>
          <div className='flex items-start justify-between gap-4'>
            <div className='space-y-1'>
              <div className='flex flex-wrap items-center gap-2'>
                <StatusBadge status={instance.status} />
                <p className='font-medium'>{target}</p>
                <Badge variant='secondary'>{instance.providerType}</Badge>
                <Badge variant='outline'>{instance.resourceType}</Badge>
              </div>
              <div className='text-sm'>
                <span className='font-mono'>{resourceName}</span>
                {!instance.resolvedName && (
                  <span className='ml-2 text-xs text-muted-foreground'>
                    (name template, not provisioned yet)
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
              {isRetryable && (
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
                className={cn(
                  'inline-flex items-center gap-1 text-sm hover:underline',
                  isPartial ? 'text-amber-700' : 'text-red-700',
                )}
              >
                {expanded ? (
                  <ChevronDown className='h-3 w-3' />
                ) : (
                  <ChevronRight className='h-3 w-3' />
                )}
                {isPartial ? 'Members that could not be added' : 'Error details'}
              </button>
              {expanded && (
                <pre
                  className={cn(
                    'mt-2 overflow-x-auto rounded p-2 text-xs',
                    isPartial ? 'bg-amber-50 text-amber-900' : 'bg-red-50 text-red-900',
                  )}
                >
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
          deleteMessage={`Delete the ${instance.resourceType} instance for ${target}?`}
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
