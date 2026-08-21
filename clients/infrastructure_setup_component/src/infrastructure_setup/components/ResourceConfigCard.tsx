import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Badge,
  Button,
  Card,
  CardContent,
  DeleteConfirmation,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Pencil, Trash2 } from 'lucide-react'
import { useState } from 'react'

import type { ResourceConfig } from '../interfaces/resourceConfig'
import { deleteResourceConfig } from '../network/mutations/deleteResourceConfig'

interface Props {
  coursePhaseID: string
  config: ResourceConfig
  onEdit: (config: ResourceConfig) => void
}

export const ResourceConfigCard = ({ coursePhaseID, config, onEdit }: Props) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [confirmOpen, setConfirmOpen] = useState(false)

  const { mutate: remove, isPending: isDeleting } = useMutation({
    mutationFn: () => deleteResourceConfig(coursePhaseID, config.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resource-configs', coursePhaseID] })
      toast({ title: 'Resource config deleted' })
      setConfirmOpen(false)
    },
    onError: (err: unknown) => {
      toast({
        title: 'Failed to delete resource config',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      })
    },
  })

  const permissionCount = Object.keys(config.permissionMapping ?? {}).length

  return (
    <>
      <Card>
        <CardContent className='flex items-start justify-between gap-4 p-4'>
          <div className='space-y-1'>
            <div className='flex flex-wrap items-center gap-2'>
              <p className='font-mono text-sm font-medium'>{config.nameTemplate}</p>
              <Badge variant='secondary'>{config.providerType}</Badge>
              <Badge variant='outline'>{config.resourceType}</Badge>
              <Badge variant='outline'>{config.scope.replace('_', ' ')}</Badge>
              {permissionCount > 0 && (
                <Badge variant='outline'>
                  {permissionCount} permission{permissionCount === 1 ? '' : 's'}
                </Badge>
              )}
            </div>
            <p className='font-mono text-xs text-muted-foreground'>{config.id}</p>
          </div>

          <div className='flex items-center gap-2'>
            <Button variant='outline' size='sm' onClick={() => onEdit(config)}>
              <Pencil className='mr-1 h-3 w-3' /> Edit
            </Button>
            <Button
              variant='outline'
              size='sm'
              onClick={() => setConfirmOpen(true)}
              disabled={isDeleting}
            >
              <Trash2 className='mr-1 h-3 w-3' /> Delete
            </Button>
          </div>
        </CardContent>
      </Card>

      {confirmOpen && (
        <DeleteConfirmation
          isOpen={confirmOpen}
          setOpen={setConfirmOpen}
          deleteMessage='Delete this resource configuration?'
          customWarning='Existing provisioned instances created from this config will be cascade-deleted.'
          onClick={(confirmed) => {
            if (confirmed) remove()
          }}
        />
      )}
    </>
  )
}

export default ResourceConfigCard
