import { coreApi } from '@core/network/api'
import { coreCache, coreKeys } from '@core/network/cache'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Button,
  DeleteConfirmation,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
  useToast,
} from '@tumaet/prompt-ui-components'
import { type FormEvent, useState } from 'react'
import type { PhaseKeyStatus } from '../interfaces/phaseKeyStatus'

interface PhaseKeyRowProps {
  phaseId: string
  phaseName: string
}

const describeKey = (status: PhaseKeyStatus | undefined, isError: boolean): string => {
  if (isError) return 'The key status could not be loaded.'
  if (!status) return 'Loading...'
  if (!status.configured) return 'No key, so AI is off for this phase.'
  const setAt = status.setAt ? new Date(status.setAt).toLocaleDateString() : 'an unknown date'
  return `Key ending in ${status.last4}, set on ${setAt}.`
}

export const PhaseKeyRow = ({ phaseId, phaseName }: PhaseKeyRowProps) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [isSetOpen, setSetOpen] = useState(false)
  const [isRemoveOpen, setRemoveOpen] = useState(false)
  const [key, setKey] = useState('')

  const { data: status, isError } = useQuery({
    queryKey: coreKeys.ai.key(phaseId),
    queryFn: () => coreApi.ai.key(phaseId),
  })

  const onKeyChanged = (title: string) => {
    coreCache.aiKeyChanged(queryClient, phaseId)
    toast({ title })
  }
  const onFailure = () =>
    toast({
      title: 'The key could not be changed',
      description: 'Please try again later!',
      variant: 'destructive',
    })

  const saveKey = useMutation({
    mutationFn: () => coreApi.ai.setKey(phaseId, key),
    onSuccess: () => {
      setSetOpen(false)
      setKey('')
      onKeyChanged('Saved the Logos key')
    },
    onError: onFailure,
  })
  const removeKey = useMutation({
    mutationFn: () => coreApi.ai.removeKey(phaseId),
    onSuccess: () => onKeyChanged('Removed the Logos key'),
    onError: onFailure,
  })

  const openSetDialog = (open: boolean) => {
    setSetOpen(open)
    if (!open) setKey('')
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    saveKey.mutate()
  }

  return (
    <fieldset aria-label={phaseName} className='flex items-center justify-between gap-4 px-3 py-4'>
      <div>
        <h3 className='text-sm font-medium'>{phaseName}</h3>
        <p className='text-sm text-muted-foreground mt-1'>{describeKey(status, isError)}</p>
      </div>
      <div className='flex gap-2'>
        <Button variant='outline' onClick={() => setSetOpen(true)} disabled={!status}>
          {status?.configured ? 'Rotate key' : 'Set key'}
        </Button>
        {status?.configured && (
          <Button variant='destructive' onClick={() => setRemoveOpen(true)}>
            Remove key
          </Button>
        )}
      </div>

      <Dialog open={isSetOpen} onOpenChange={openSetDialog}>
        <DialogContent>
          <form onSubmit={submit} className='space-y-4'>
            <DialogHeader>
              <DialogTitle>Logos key for {phaseName}</DialogTitle>
              <DialogDescription>
                The key is stored encrypted and never shown again. Enter the same key in several
                phases to share one budget.
              </DialogDescription>
            </DialogHeader>
            <div className='space-y-2'>
              <Label htmlFor={`logos-key-${phaseId}`}>Logos key</Label>
              <Input
                id={`logos-key-${phaseId}`}
                type='password'
                autoComplete='off'
                value={key}
                onChange={(event) => setKey(event.target.value)}
              />
            </div>
            <DialogFooter>
              <Button type='submit' disabled={key.trim().length < 8 || saveKey.isPending}>
                Save key
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {isRemoveOpen && (
        <DeleteConfirmation
          isOpen={isRemoveOpen}
          setOpen={setRemoveOpen}
          deleteMessage={`Remove the Logos key of ${phaseName}?`}
          customWarning='AI is turned off for this phase until a key is set again. Recorded AI calls stay.'
          onClick={(confirmed) => confirmed && removeKey.mutate()}
        />
      )}
    </fieldset>
  )
}
