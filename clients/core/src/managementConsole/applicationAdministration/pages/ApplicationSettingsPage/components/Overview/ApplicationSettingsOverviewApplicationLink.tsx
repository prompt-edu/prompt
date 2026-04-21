import { useState } from 'react'
import { Copy, Check, Link } from 'lucide-react'
import { Button } from '@tumaet/prompt-ui-components'
import { useApplicationStore } from '../../../../zustand/useApplicationStore'

export const ApplicationSettingsOverviewApplicationLink = () => {
  const { coursePhase } = useApplicationStore()
  const [copied, setCopied] = useState(false)

  const applicationLink = coursePhase?.id
    ? `${window.location.origin}/apply/${coursePhase.id}`
    : null

  const handleCopy = () => {
    if (!applicationLink) return
    navigator.clipboard.writeText(applicationLink)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className='w-full p-3 border rounded-md border-blue-200'>
      <div className='flex items-center space-x-2'>
        <Link className='h-4 w-4 text-blue-500 shrink-0' />
        <span className='text-sm text-secondary-foreground font-medium'>Application Link</span>
      </div>
      <div className='flex items-center mt-2 gap-2'>
        <span className='text-sm text-muted-foreground truncate flex-1 font-mono bg-muted px-2 py-1 rounded'>
          {applicationLink ?? 'Not available'}
        </span>
        <Button
          variant='outline'
          size='sm'
          onClick={handleCopy}
          disabled={!applicationLink}
          className='shrink-0'
        >
          {copied ? <Check className='h-4 w-4 text-green-500' /> : <Copy className='h-4 w-4' />}
          {copied ? 'Copied!' : 'Copy'}
        </Button>
      </div>
    </div>
  )
}
