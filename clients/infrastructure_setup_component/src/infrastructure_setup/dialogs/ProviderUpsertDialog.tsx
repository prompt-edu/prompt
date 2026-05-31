import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  useToast,
} from '@tumaet/prompt-ui-components'

import {
  AuthField,
  ProviderConfig,
  ProviderType,
  providerTypes,
} from '../interfaces/providerConfig'
import { getProviderAuthFields } from '../network/queries/getProviderAuthFields'
import { upsertProviderConfig } from '../network/mutations/upsertProviderConfig'

interface Props {
  coursePhaseID: string
  open: boolean
  onOpenChange: (open: boolean) => void
  // When editing an existing provider, pass it in; the type is then fixed.
  // When creating, pass undefined and provide the list of already-configured types
  // so the type selector can exclude duplicates.
  existingProvider?: ProviderConfig
  configuredTypes: ProviderType[]
}

export const ProviderUpsertDialog = ({
  coursePhaseID,
  open,
  onOpenChange,
  existingProvider,
  configuredTypes,
}: Props) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()

  const availableTypes = useMemo(
    () => providerTypes.filter((t) => !configuredTypes.includes(t)),
    [configuredTypes],
  )

  const [selectedType, setSelectedType] = useState<ProviderType | undefined>(
    existingProvider?.providerType ?? availableTypes[0],
  )
  const [values, setValues] = useState<Record<string, string>>({})

  useEffect(() => {
    if (!open) return
    setSelectedType(existingProvider?.providerType ?? availableTypes[0])
    setValues({})
  }, [open, existingProvider, availableTypes])

  const {
    data: authFields,
    isLoading: fieldsLoading,
    isError: fieldsError,
  } = useQuery({
    queryKey: ['provider-auth-fields', coursePhaseID, selectedType],
    queryFn: () => getProviderAuthFields(coursePhaseID, selectedType!),
    enabled: open && !!selectedType,
  })

  const { mutate: save, isPending: isSaving } = useMutation({
    mutationFn: () =>
      upsertProviderConfig(coursePhaseID, {
        providerType: selectedType!,
        credentials: values,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['provider-configs', coursePhaseID] })
      toast({
        title: existingProvider ? 'Provider updated' : 'Provider added',
        description: `Credentials for ${selectedType} saved.`,
      })
      onOpenChange(false)
    },
    onError: (err: unknown) => {
      toast({
        title: 'Failed to save provider',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      })
    },
  })

  const requiredFilled =
    !!authFields && authFields.every((f) => !f.required || (values[f.name] ?? '').trim() !== '')

  const noTypesAvailable = !existingProvider && availableTypes.length === 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            if (!selectedType || !requiredFilled || isSaving) return
            save()
          }}
        >
          <DialogHeader>
            <DialogTitle>
              {existingProvider ? `Edit ${existingProvider.providerType} provider` : 'Add provider'}
            </DialogTitle>
            <DialogDescription>
              {existingProvider
                ? 'Credentials are write-only — re-enter all values to overwrite the stored credentials.'
                : 'Configure credentials for a new infrastructure provider.'}
            </DialogDescription>
          </DialogHeader>

          <div className='space-y-4 py-4'>
            {!existingProvider && (
              <div className='space-y-1'>
                <Label htmlFor='providerType'>Provider type</Label>
                {noTypesAvailable ? (
                  <p className='text-sm text-muted-foreground'>
                    All five providers are already configured.
                  </p>
                ) : (
                  <Select
                    value={selectedType}
                    onValueChange={(v) => setSelectedType(v as ProviderType)}
                  >
                    <SelectTrigger id='providerType'>
                      <SelectValue placeholder='Select a provider' />
                    </SelectTrigger>
                    <SelectContent>
                      {availableTypes.map((t) => (
                        <SelectItem key={t} value={t}>
                          {t}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </div>
            )}

            {selectedType && fieldsLoading && (
              <p className='text-sm text-muted-foreground'>Loading credential fields…</p>
            )}
            {selectedType && fieldsError && (
              <p className='text-sm text-red-600'>Failed to load credential fields.</p>
            )}
            {selectedType &&
              authFields?.map((field: AuthField) => (
                <div key={field.name} className='space-y-1'>
                  <Label htmlFor={field.name}>
                    {field.label}
                    {field.required && <span className='ml-1 text-red-500'>*</span>}
                  </Label>
                  <Input
                    id={field.name}
                    type={field.type === 'password' ? 'password' : 'text'}
                    value={values[field.name] ?? ''}
                    onChange={(e) =>
                      setValues((prev) => ({ ...prev, [field.name]: e.target.value }))
                    }
                    autoComplete='off'
                  />
                  {field.description && (
                    <p className='text-xs text-muted-foreground'>{field.description}</p>
                  )}
                </div>
              ))}
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
              disabled={isSaving}
            >
              Cancel
            </Button>
            <Button
              type='submit'
              disabled={!selectedType || !requiredFilled || isSaving || noTypesAvailable}
            >
              {isSaving ? 'Saving…' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default ProviderUpsertDialog
