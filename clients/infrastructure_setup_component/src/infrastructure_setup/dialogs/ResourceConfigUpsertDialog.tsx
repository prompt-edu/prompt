import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
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
  RadioGroup,
  RadioGroupItem,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Textarea,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Plus, Trash2 } from 'lucide-react'

import { ProviderType } from '../interfaces/providerConfig'
import {
  CreateResourceConfigRequest,
  ResourceConfig,
  Scope,
  UpdateResourceConfigRequest,
} from '../interfaces/resourceConfig'
import { createResourceConfig } from '../network/mutations/createResourceConfig'
import { updateResourceConfig } from '../network/mutations/updateResourceConfig'

interface Props {
  coursePhaseID: string
  open: boolean
  onOpenChange: (open: boolean) => void
  existing?: ResourceConfig
  // Provider types the course phase has credentials for — used as the options for the
  // provider select when creating a new resource config.
  availableProviderTypes: ProviderType[]
}

interface RoleRow {
  id: number
  role: string
  permission: string
}

const toRoleRows = (mapping: Record<string, string> | undefined): RoleRow[] =>
  Object.entries(mapping ?? {}).map(([role, permission], idx) => ({
    id: idx,
    role,
    permission,
  }))

const fromRoleRows = (rows: RoleRow[]): Record<string, string> => {
  const out: Record<string, string> = {}
  for (const row of rows) {
    const role = row.role.trim()
    if (role) out[role] = row.permission
  }
  return out
}

export const ResourceConfigUpsertDialog = ({
  coursePhaseID,
  open,
  onOpenChange,
  existing,
  availableProviderTypes,
}: Props) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()

  const initialProviderType = useMemo<ProviderType | undefined>(
    () => existing?.providerType ?? availableProviderTypes[0],
    [existing, availableProviderTypes],
  )

  const [providerType, setProviderType] = useState<ProviderType | undefined>(initialProviderType)
  const [resourceType, setResourceType] = useState(existing?.resourceType ?? '')
  const [scope, setScope] = useState<Scope>(existing?.scope ?? 'per_team')
  const [nameTemplate, setNameTemplate] = useState(existing?.nameTemplate ?? '')
  const [permissionRows, setPermissionRows] = useState<RoleRow[]>(
    toRoleRows(existing?.permissionMapping),
  )
  const [extraConfigRaw, setExtraConfigRaw] = useState(
    JSON.stringify(existing?.resourceExtraConfig ?? {}, null, 2),
  )
  const [extraConfigError, setExtraConfigError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setProviderType(existing?.providerType ?? availableProviderTypes[0])
    setResourceType(existing?.resourceType ?? '')
    setScope(existing?.scope ?? 'per_team')
    setNameTemplate(existing?.nameTemplate ?? '')
    setPermissionRows(toRoleRows(existing?.permissionMapping))
    setExtraConfigRaw(JSON.stringify(existing?.resourceExtraConfig ?? {}, null, 2))
    setExtraConfigError(null)
  }, [open, existing, availableProviderTypes])

  const parseExtraConfig = (): Record<string, unknown> | null => {
    const raw = extraConfigRaw.trim()
    if (raw === '') return {}
    try {
      const parsed = JSON.parse(raw)
      if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
        setExtraConfigError('Must be a JSON object.')
        return null
      }
      setExtraConfigError(null)
      return parsed
    } catch (err) {
      setExtraConfigError(err instanceof Error ? err.message : 'Invalid JSON.')
      return null
    }
  }

  const { mutate: save, isPending } = useMutation({
    mutationFn: async () => {
      const extra = parseExtraConfig()
      if (extra === null) throw new Error('Invalid extra config JSON')
      const permissionMapping = fromRoleRows(permissionRows)

      if (existing) {
        const req: UpdateResourceConfigRequest = {
          resourceType: resourceType.trim(),
          scope,
          nameTemplate: nameTemplate.trim(),
          permissionMapping,
          resourceExtraConfig: extra,
        }
        return updateResourceConfig(coursePhaseID, existing.id, req)
      }

      if (!providerType) throw new Error('Provider type required')
      const req: CreateResourceConfigRequest = {
        providerType,
        resourceType: resourceType.trim(),
        scope,
        nameTemplate: nameTemplate.trim(),
        permissionMapping,
        resourceExtraConfig: extra,
      }
      return createResourceConfig(coursePhaseID, req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resource-configs', coursePhaseID] })
      toast({
        title: existing ? 'Resource config updated' : 'Resource config created',
      })
      onOpenChange(false)
    },
    onError: (err: unknown) => {
      toast({
        title: 'Failed to save resource config',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      })
    },
  })

  const canSubmit =
    !!resourceType.trim() &&
    !!nameTemplate.trim() &&
    !extraConfigError &&
    (existing ? true : !!providerType) &&
    !isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-2xl'>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            if (!canSubmit) return
            save()
          }}
        >
          <DialogHeader>
            <DialogTitle>
              {existing ? 'Edit resource configuration' : 'New resource configuration'}
            </DialogTitle>
            <DialogDescription>
              Resource configurations describe what to provision per team or per student during
              execution.
            </DialogDescription>
          </DialogHeader>

          <div className='space-y-4 py-4'>
            <div className='space-y-1'>
              <Label htmlFor='providerType'>Provider</Label>
              {existing ? (
                <Input id='providerType' value={existing.providerType} disabled />
              ) : availableProviderTypes.length === 0 ? (
                <p className='text-sm text-red-600'>
                  No providers configured yet. Add a provider before creating a resource config.
                </p>
              ) : (
                <Select
                  value={providerType}
                  onValueChange={(v) => setProviderType(v as ProviderType)}
                >
                  <SelectTrigger id='providerType'>
                    <SelectValue placeholder='Select a provider' />
                  </SelectTrigger>
                  <SelectContent>
                    {availableProviderTypes.map((t) => (
                      <SelectItem key={t} value={t}>
                        {t}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </div>

            <div className='space-y-1'>
              <Label htmlFor='resourceType'>Resource type</Label>
              <Input
                id='resourceType'
                value={resourceType}
                onChange={(e) => setResourceType(e.target.value)}
                placeholder='group, channel, collection, project, …'
              />
              <p className='text-xs text-muted-foreground'>
                Provider-specific resource kind (e.g. <code>group</code> for GitLab,{' '}
                <code>channel</code> for Slack).
              </p>
            </div>

            <div className='space-y-1'>
              <Label>Scope</Label>
              <RadioGroup
                value={scope}
                onValueChange={(v) => setScope(v as Scope)}
                className='flex gap-6'
              >
                <Label className='flex items-center gap-2 font-normal'>
                  <RadioGroupItem value='per_team' /> per team
                </Label>
                <Label className='flex items-center gap-2 font-normal'>
                  <RadioGroupItem value='per_student' /> per student
                </Label>
              </RadioGroup>
            </div>

            <div className='space-y-1'>
              <Label htmlFor='nameTemplate'>Name template</Label>
              <Input
                id='nameTemplate'
                value={nameTemplate}
                onChange={(e) => setNameTemplate(e.target.value)}
                placeholder='{{semesterTag}}-{{teamName}}'
                className='font-mono'
              />
              <p className='text-xs text-muted-foreground'>
                Available variables: <code>{`{{semesterTag}}`}</code>, <code>{`{{teamName}}`}</code>
                , <code>{`{{studentFirstName}}`}</code>, <code>{`{{studentLastName}}`}</code>,{' '}
                <code>{`{{studentEmail}}`}</code>.
              </p>
            </div>

            <div className='space-y-1'>
              <div className='flex items-center justify-between'>
                <Label>Permission mapping</Label>
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  onClick={() =>
                    setPermissionRows((prev) => [
                      ...prev,
                      { id: Date.now(), role: '', permission: '' },
                    ])
                  }
                >
                  <Plus className='mr-1 h-3 w-3' /> Add row
                </Button>
              </div>
              <p className='text-xs text-muted-foreground'>
                Maps a logical role (e.g. <code>student</code>, <code>tutor</code>) to a
                provider-specific permission level.
              </p>
              {permissionRows.length === 0 ? (
                <p className='text-sm text-muted-foreground'>No permissions mapped.</p>
              ) : (
                <div className='space-y-2'>
                  {permissionRows.map((row, idx) => (
                    <div key={row.id} className='flex items-center gap-2'>
                      <Input
                        value={row.role}
                        onChange={(e) =>
                          setPermissionRows((prev) =>
                            prev.map((r, i) => (i === idx ? { ...r, role: e.target.value } : r)),
                          )
                        }
                        placeholder='role (e.g. student)'
                      />
                      <Input
                        value={row.permission}
                        onChange={(e) =>
                          setPermissionRows((prev) =>
                            prev.map((r, i) =>
                              i === idx ? { ...r, permission: e.target.value } : r,
                            ),
                          )
                        }
                        placeholder='permission (e.g. developer)'
                      />
                      <Button
                        type='button'
                        variant='ghost'
                        size='icon'
                        onClick={() =>
                          setPermissionRows((prev) => prev.filter((_, i) => i !== idx))
                        }
                        aria-label='Remove row'
                      >
                        <Trash2 className='h-4 w-4' />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className='space-y-1'>
              <Label htmlFor='extraConfig'>Resource extra config (JSON)</Label>
              <Textarea
                id='extraConfig'
                value={extraConfigRaw}
                onChange={(e) => {
                  setExtraConfigRaw(e.target.value)
                  setExtraConfigError(null)
                }}
                onBlur={() => parseExtraConfig()}
                rows={6}
                className='font-mono text-xs'
              />
              {extraConfigError && (
                <p className='text-sm text-red-600'>JSON error: {extraConfigError}</p>
              )}
              <p className='text-xs text-muted-foreground'>
                Provider-specific settings (e.g. Rancher <code>roleTemplateId</code>). Leave as{' '}
                <code>{`{}`}</code> if not needed.
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type='submit' disabled={!canSubmit}>
              {isPending ? 'Saving…' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default ResourceConfigUpsertDialog
