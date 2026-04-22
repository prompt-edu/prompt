import { useState } from 'react'
import { Copy, Check, Link } from 'lucide-react'
import { Button } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const SurveyLinkCard = () => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const [copied, setCopied] = useState(false)

  const surveyLink =
    courseId && phaseId
      ? `${window.location.origin}/management/course/${courseId}/${phaseId}`
      : null

  const handleCopy = () => {
    if (!surveyLink) return
    navigator.clipboard.writeText(surveyLink)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className='w-full p-3 border rounded-md border-blue-200'>
      <div className='flex items-center space-x-2'>
        <Link className='h-4 w-4 text-blue-500 shrink-0' />
        <span className='text-sm text-secondary-foreground font-medium'>Survey Link</span>
      </div>
      <div className='flex items-center mt-2 gap-2'>
        <span className='text-sm text-muted-foreground truncate flex-1 font-mono bg-muted px-2 py-1 rounded'>
          {surveyLink ?? 'Not available'}
        </span>
        <Button
          variant='outline'
          size='sm'
          onClick={handleCopy}
          disabled={!surveyLink}
          className='shrink-0'
        >
          {copied ? <Check className='h-4 w-4 text-green-500' /> : <Copy className='h-4 w-4' />}
          {copied ? 'Copied!' : 'Copy'}
        </Button>
      </div>
    </div>
  )
}
