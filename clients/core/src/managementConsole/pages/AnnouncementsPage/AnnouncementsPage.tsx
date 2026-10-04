import { useNow } from '@core/announcementBanner/hooks/useNow'
import type { Announcement, UpsertAnnouncement } from '@core/interfaces/announcement'
import {
  useAnnouncements,
  useCreateAnnouncement,
  useDeleteAnnouncement,
  useUpdateAnnouncement,
} from '@core/network/hooks/useAnnouncements'
import { Button, Label, PromptTable, Switch } from '@tumaet/prompt-ui-components'
import { Plus } from 'lucide-react'
import { useCallback, useMemo, useState } from 'react'
import { AnnouncementFormDialog } from './components/AnnouncementFormDialog'
import { getAnnouncementTableActions } from './components/announcementTableActions'
import { getAnnouncementTableColumns } from './components/announcementTableColumns'

const toUpsert = ({
  severity,
  title,
  message,
  linkUrl,
  linkLabel,
  startsAt,
  expiresAt,
  enabled,
}: Announcement): UpsertAnnouncement => ({
  severity,
  title,
  message,
  linkUrl,
  linkLabel,
  startsAt,
  expiresAt,
  enabled,
})

export const AnnouncementsPage = () => {
  const [showExpired, setShowExpired] = useState(false)
  const { data: announcements = [], isLoading, isError } = useAnnouncements(showExpired)
  const createAnnouncement = useCreateAnnouncement()
  const updateAnnouncement = useUpdateAnnouncement()
  const deleteAnnouncement = useDeleteAnnouncement()

  const [createOpen, setCreateOpen] = useState(false)
  const [editAnnouncement, setEditAnnouncement] = useState<Announcement | null>(null)

  const handleCreate = (values: UpsertAnnouncement) => {
    createAnnouncement.mutate(values, { onSuccess: () => setCreateOpen(false) })
  }

  const handleUpdate = (values: UpsertAnnouncement) => {
    if (!editAnnouncement) return
    updateAnnouncement.mutate(
      { id: editAnnouncement.id, announcement: values },
      { onSuccess: () => setEditAnnouncement(null) },
    )
  }

  const { mutate: updateMutate } = updateAnnouncement
  const handleToggleEnabled = useCallback(
    (announcement: Announcement, enabled: boolean) => {
      updateMutate({ id: announcement.id, announcement: { ...toUpsert(announcement), enabled } })
    },
    [updateMutate],
  )

  const { mutate: deleteMutate } = deleteAnnouncement
  const handleDelete = useCallback(
    (selected: Announcement[]) => {
      for (const announcement of selected) {
        deleteMutate(announcement.id)
      }
    },
    [deleteMutate],
  )

  const now = useNow()
  const columns = useMemo(
    () => getAnnouncementTableColumns(handleToggleEnabled, now),
    [handleToggleEnabled, now],
  )
  const actions = useMemo(
    () => getAnnouncementTableActions({ onEdit: setEditAnnouncement, onDelete: handleDelete }),
    [handleDelete],
  )

  return (
    <div className='flex w-full flex-col gap-6'>
      <div className='flex flex-wrap items-center justify-between gap-4'>
        <div className='flex flex-col gap-1'>
          <h1 className='text-3xl font-bold tracking-tight'>Announcements</h1>
          <p className='text-muted-foreground'>
            Show announcement banners, e.g. for maintenance windows or outages, to all users.
          </p>
        </div>
        <div className='flex items-center gap-4'>
          <div className='flex items-center gap-2'>
            <Switch id='show-expired' checked={showExpired} onCheckedChange={setShowExpired} />
            <Label htmlFor='show-expired'>Show expired</Label>
          </div>
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className='mr-2 h-4 w-4' />
            Create Announcement
          </Button>
        </div>
      </div>

      {isLoading && <p className='text-sm text-muted-foreground'>Loading announcements…</p>}
      {isError && <p className='text-sm text-destructive'>Failed to load announcements.</p>}
      {!isLoading && !isError && (
        <PromptTable
          data={announcements}
          columns={columns}
          actions={actions}
          onRowClick={setEditAnnouncement}
        />
      )}

      {createOpen && (
        <AnnouncementFormDialog
          open={createOpen}
          onClose={() => setCreateOpen(false)}
          onSubmit={handleCreate}
          isPending={createAnnouncement.isPending}
        />
      )}

      {editAnnouncement && (
        <AnnouncementFormDialog
          open={!!editAnnouncement}
          onClose={() => setEditAnnouncement(null)}
          announcement={editAnnouncement}
          onSubmit={handleUpdate}
          isPending={updateAnnouncement.isPending}
        />
      )}
    </div>
  )
}
