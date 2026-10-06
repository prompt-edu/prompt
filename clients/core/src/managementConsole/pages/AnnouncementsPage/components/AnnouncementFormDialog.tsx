import { AnnouncementBanner } from '@core/announcementBanner/AnnouncementBanner'
import type {
  Announcement,
  AnnouncementSeverity,
  UpsertAnnouncement,
} from '@core/interfaces/announcement'
import {
  Button,
  DatePicker,
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
  Switch,
  Textarea,
} from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import { SEVERITY_LABELS } from './announcementLabels'
import { validateAnnouncementForm } from './validateAnnouncementForm'

interface AnnouncementFormDialogProps {
  open: boolean
  onClose: () => void
  announcement?: Announcement
  onSubmit: (values: UpsertAnnouncement) => void
  isPending: boolean
}

const MAX_TITLE_LENGTH = 100
const MAX_MESSAGE_LENGTH = 300
const MAX_LINK_LABEL_LENGTH = 50
const TIME_ZONE = Intl.DateTimeFormat().resolvedOptions().timeZone

const toDate = (value: string | null | undefined): Date | undefined =>
  value ? new Date(value) : undefined

export const AnnouncementFormDialog = ({
  open,
  onClose,
  announcement,
  onSubmit,
  isPending,
}: AnnouncementFormDialogProps) => {
  const [severity, setSeverity] = useState<AnnouncementSeverity>(announcement?.severity ?? 'info')
  const [title, setTitle] = useState(announcement?.title ?? '')
  const [message, setMessage] = useState(announcement?.message ?? '')
  const [linkUrl, setLinkUrl] = useState(announcement?.linkUrl ?? '')
  const [linkLabel, setLinkLabel] = useState(announcement?.linkLabel ?? '')
  const [startsAt, setStartsAt] = useState(toDate(announcement?.startsAt))
  const [expiresAt, setExpiresAt] = useState(toDate(announcement?.expiresAt))
  const [enabled, setEnabled] = useState(announcement?.enabled ?? false)

  const values: UpsertAnnouncement = {
    severity,
    title: title.trim(),
    message: message.trim(),
    linkUrl: linkUrl.trim(),
    linkLabel: linkLabel.trim(),
    startsAt: startsAt?.toISOString() ?? null,
    expiresAt: expiresAt?.toISOString() ?? null,
    enabled,
  }
  const errors = validateAnnouncementForm(values)
  const isValid = Object.keys(errors).length === 0

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) onClose()
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className='sm:max-w-2xl max-h-[90vh] overflow-y-auto'>
        <DialogHeader>
          <DialogTitle>{announcement ? 'Edit Announcement' : 'Create Announcement'}</DialogTitle>
          <DialogDescription>
            Announcement banners are shown at the top of every page, including the public
            application pages.
          </DialogDescription>
        </DialogHeader>

        <section className='rounded-lg border-2 border-dashed border-primary/40 bg-muted/40 p-3'>
          <p className='mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground'>
            Preview – only you can see this
          </p>
          <div className='overflow-hidden rounded-md border'>
            <AnnouncementBanner
              severity={severity}
              title={values.title}
              message={values.message || 'Your message will appear here.'}
              linkUrl={errors.linkUrl ? '' : values.linkUrl}
              linkLabel={values.linkLabel}
            />
          </div>
        </section>

        <div className='grid gap-4 py-2 sm:grid-cols-2'>
          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='announcement-severity'>Severity</Label>
            <Select
              value={severity}
              onValueChange={(value) => setSeverity(value as AnnouncementSeverity)}
            >
              <SelectTrigger id='announcement-severity'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(SEVERITY_LABELS).map(([value, label]) => (
                  <SelectItem key={value} value={value}>
                    {label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='announcement-title'>Title (optional)</Label>
            <Input
              id='announcement-title'
              value={title}
              maxLength={MAX_TITLE_LENGTH}
              onChange={(e) => setTitle(e.target.value)}
              placeholder='Scheduled maintenance'
            />
          </div>

          <div className='flex flex-col gap-1.5 sm:col-span-2'>
            <div className='flex items-center justify-between'>
              <Label htmlFor='announcement-message'>Message</Label>
              <span className='text-xs text-muted-foreground'>
                {message.length}/{MAX_MESSAGE_LENGTH}
              </span>
            </div>
            <Textarea
              id='announcement-message'
              value={message}
              maxLength={MAX_MESSAGE_LENGTH}
              onChange={(e) => setMessage(e.target.value)}
              placeholder='PROMPT will be unavailable on Saturday from 08:00 to 10:00.'
              rows={3}
            />
          </div>

          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='announcement-link-url'>Link URL (optional)</Label>
            <Input
              id='announcement-link-url'
              value={linkUrl}
              onChange={(e) => setLinkUrl(e.target.value)}
              placeholder='https://'
            />
            {errors.linkUrl && <p className='text-sm text-destructive'>{errors.linkUrl}</p>}
          </div>

          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='announcement-link-label'>Link label (optional)</Label>
            <Input
              id='announcement-link-label'
              value={linkLabel}
              maxLength={MAX_LINK_LABEL_LENGTH}
              onChange={(e) => setLinkLabel(e.target.value)}
              placeholder='Learn more'
            />
            {errors.linkLabel && <p className='text-sm text-destructive'>{errors.linkLabel}</p>}
          </div>

          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='announcement-starts-at'>Starts (optional)</Label>
            <DatePicker
              id='announcement-starts-at'
              date={startsAt}
              onSelect={setStartsAt}
              withTime
              className='w-full'
              placeholder='Immediately'
            />
          </div>

          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='announcement-expires-at'>Expires (optional)</Label>
            <DatePicker
              id='announcement-expires-at'
              date={expiresAt}
              onSelect={setExpiresAt}
              withTime
              className='w-full'
              placeholder='Never'
            />
            {errors.expiresAt && <p className='text-sm text-destructive'>{errors.expiresAt}</p>}
          </div>

          <p className='text-xs text-muted-foreground sm:col-span-2'>
            Times are in your local time zone ({TIME_ZONE}).
          </p>

          <div className='flex items-center justify-between gap-4 rounded-lg border p-3 sm:col-span-2'>
            <div className='flex flex-col'>
              <Label htmlFor='announcement-enabled'>Enabled</Label>
              <span className='text-sm text-muted-foreground'>
                Only enabled announcements are shown to users within their schedule.
              </span>
            </div>
            <Switch id='announcement-enabled' checked={enabled} onCheckedChange={setEnabled} />
          </div>
        </div>

        <DialogFooter>
          <Button variant='outline' onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={() => onSubmit(values)} disabled={!isValid || isPending}>
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
