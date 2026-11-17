import type { Org } from '@core/interfaces/org'
import { useDeleteOrg } from '@core/network/hooks/useOrgs'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  DeleteConfirmation,
} from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { orgLabel } from '../utils/orgHierarchy'

interface OrgDangerZoneProps {
  org: Org
}

export const OrgDangerZone = ({ org }: OrgDangerZoneProps) => {
  const navigate = useNavigate()
  const deleteOrg = useDeleteOrg()
  const [confirmOpen, setConfirmOpen] = useState(false)

  const handleDelete = (confirmed: boolean) => {
    if (confirmed) {
      deleteOrg.mutate(org.id, { onSuccess: () => navigate('/management/admin/orgs') })
    }
    setConfirmOpen(false)
  }

  return (
    <>
      <Card className='border-destructive/50'>
        <CardHeader>
          <CardTitle>Danger Zone</CardTitle>
          <CardDescription>Actions that cannot be undone.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className='flex items-center justify-between gap-4 rounded-md border px-3 py-4'>
            <div>
              <h3 className='text-sm font-medium'>Delete Org</h3>
              <p className='mt-1 text-sm text-muted-foreground'>
                Deletes the org together with its Keycloak roles and every membership. An org that
                still has child orgs or courses cannot be deleted.
              </p>
            </div>
            <Button
              variant='destructive'
              disabled={deleteOrg.isPending}
              onClick={() => setConfirmOpen(true)}
            >
              Delete
            </Button>
          </div>
        </CardContent>
      </Card>
      <DeleteConfirmation
        isOpen={confirmOpen}
        setOpen={setConfirmOpen}
        deleteMessage={`Are you sure you want to delete ${orgLabel(org)}?`}
        onClick={handleDelete}
      />
    </>
  )
}
