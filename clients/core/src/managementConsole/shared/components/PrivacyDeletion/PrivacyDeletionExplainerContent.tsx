import { Mail, Recycle, Trash2 } from 'lucide-react'
import { type ReactNode } from 'react'
import { PrivacyServiceAvailability } from '../Privacy/PrivacyServiceAvailability'

export type PrivacyDeletionVariant = 'review' | 'initiate'

interface Bullet {
  icon: ReactNode
  text: ReactNode
}

const bulletIcon = (Icon: typeof Trash2) => (
  <Icon className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
)

const howItWorks: Record<PrivacyDeletionVariant, Bullet[]> = {
  review: [
    {
      icon: bulletIcon(Trash2),
      text: 'Approval starts the deletion immediately across all services. This action cannot be undone.',
    },
    {
      icon: bulletIcon(Recycle),
      text: <>Rejection leaves the user&apos;s data unchanged.</>,
    },
    {
      icon: bulletIcon(Mail),
      text: 'The user is notified of your decision and may submit another deletion request later.',
    },
  ],
  initiate: [
    {
      icon: bulletIcon(Trash2),
      text: 'Starting the deletion runs it immediately across all services. This action cannot be undone.',
    },
    {
      icon: bulletIcon(Recycle),
      text: 'Closing the dialog before all batches complete will stop further batches but not roll back ones already submitted.',
    },
    {
      icon: bulletIcon(Mail),
      text: 'Affected users with a Keycloak account are notified that their data was deleted.',
    },
  ],
}

export function PrivacyDeletionHowItWorks({ variant }: { variant: PrivacyDeletionVariant }) {
  return (
    <div className='flex flex-col gap-3'>
      <p className='font-medium'>How it works</p>
      <ul className='flex flex-col gap-2.5'>
        {howItWorks[variant].map((b, i) => (
          <li key={i} className='flex items-start gap-3'>
            {b.icon}
            <span>{b.text}</span>
          </li>
        ))}
      </ul>

      <PrivacyServiceAvailability />
    </div>
  )
}

export function PrivacyDeletionWhatGetsDeleted({ variant }: { variant: PrivacyDeletionVariant }) {
  return (
    <div className='flex flex-col gap-3'>
      <p className='font-medium'>What gets deleted</p>
      {variant === 'review' ? (
        <p className='text-muted-foreground'>
          Approval permanently removes the user&apos;s personal data. Only services the user has
          actually interacted with are contacted, the rest are skipped. This may include (depending
          on the user&apos;s history):
        </p>
      ) : (
        <p className='text-muted-foreground'>
          Permanently removes each student&apos;s personal data. Only services the student has
          actually interacted with are contacted, the rest are skipped. This may include (depending
          on the student&apos;s history):
        </p>
      )}
      <ul className='list-disc pl-5 text-muted-foreground space-y-1'>
        <li>Course enrollments</li>
        <li>Application data</li>
        <li>Assessment results</li>
        <li>Team allocation</li>
        <li>Instructor notes</li>
      </ul>
    </div>
  )
}
