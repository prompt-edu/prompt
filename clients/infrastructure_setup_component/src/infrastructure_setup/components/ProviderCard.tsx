import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Badge,
  Button,
  Card,
  CardContent,
  DeleteConfirmation,
  useToast,
} from '@tumaet/prompt-ui-components'
import axios from 'axios'
import { CheckCircle2, Pencil, ShieldCheck, Trash2 } from 'lucide-react'
import { useState } from 'react'

import type { ProviderConfig } from '../interfaces/providerConfig'
import { deleteProviderConfig } from '../network/mutations/deleteProviderConfig'
import { validateProviderConfig } from '../network/mutations/validateProviderConfig'

interface Props {
  coursePhaseID: string
  provider: ProviderConfig
  onEdit: (provider: ProviderConfig) => void
}

export const ProviderCard = ({ coursePhaseID, provider, onEdit }: Props) => {
  const { toast } = useToast()
  const queryClient = useQueryClient()
  const [confirmOpen, setConfirmOpen] = useState(false)

  const describeError = (err: unknown, fallback: string) =>
    axios.isAxiosError(err) && err.response?.data?.error
      ? String(err.response.data.error)
      : err instanceof Error
        ? err.message
        : fallback

  const { mutate: validate, isPending: isValidating } = useMutation({
    mutationFn: () => validateProviderConfig(coursePhaseID, provider.providerType),
    onSuccess: () => {
      toast({
        title: 'Credentials valid',
        description: `${provider.providerType} credentials accepted by the provider.`,
      })
    },
    onError: (err: unknown) => {
      toast({
        title: 'Validation failed',
        description: describeError(err, 'Validation failed.'),
        variant: 'destructive',
      })
    },
  })

  const { mutate: remove, isPending: isDeleting } = useMutation({
    mutationFn: () => deleteProviderConfig(coursePhaseID, provider.providerType),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['provider-configs', coursePhaseID] })
      // Resource configs cascade-delete when a provider is removed; refresh that list too.
      queryClient.invalidateQueries({ queryKey: ['resource-configs', coursePhaseID] })
      queryClient.invalidateQueries({ queryKey: ['instances', coursePhaseID] })
      toast({ title: `${provider.providerType} provider removed` })
      setConfirmOpen(false)
    },
    onError: (err: unknown) => {
      toast({
        title: 'Failed to delete provider',
        description: describeError(err, 'Unknown error'),
        variant: 'destructive',
      })
    },
  })

  return (
    <>
      <Card>
        <CardContent className='flex items-center justify-between gap-4 p-4'>
          <div className='flex items-center gap-3'>
            <ShieldCheck className='h-5 w-5 text-blue-500' />
            <div>
              <div className='flex items-center gap-2'>
                <p className='font-medium capitalize'>{provider.providerType}</p>
                <Badge variant='secondary' className='gap-1'>
                  <CheckCircle2 className='h-3 w-3' /> configured
                </Badge>
              </div>
              <p className='font-mono text-xs text-muted-foreground'>{provider.id}</p>
            </div>
          </div>

          <div className='flex items-center gap-2'>
            <Button variant='outline' size='sm' onClick={() => validate()} disabled={isValidating}>
              {isValidating ? 'Validating…' : 'Validate'}
            </Button>
            <Button variant='outline' size='sm' onClick={() => onEdit(provider)}>
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
          deleteMessage={`Remove the ${provider.providerType} provider?`}
          customWarning={
            'This also deletes every resource configuration and resource instance ' +
            'that uses this provider for this course phase. Provisioned external ' +
            'resources (groups, channels, …) are NOT touched.'
          }
          onClick={(confirmed) => {
            if (confirmed) remove()
          }}
        />
      )}
    </>
  )
}

export default ProviderCard
