import type { ReactNode } from 'react'

interface Props {
  // The anchor the provisioning checklist links to.
  id: string
  step: number
  title: string
  description: ReactNode
  action?: ReactNode
  children: ReactNode
}

export const ConfigurationSection = ({ id, step, title, description, action, children }: Props) => (
  <section id={id} className='scroll-mt-6 space-y-3'>
    <div className='flex items-start justify-between gap-4'>
      <div className='space-y-1'>
        <h2 className='text-xl font-semibold'>
          {step}. {title}
        </h2>
        <p className='text-sm text-muted-foreground'>{description}</p>
      </div>
      {action}
    </div>
    {children}
  </section>
)
