import { useGetMailingIsConfigured, useModifyCoursePhase } from '@tumaet/prompt-shared-state'
import {
  AvailableMailPlaceholders,
  availablePlaceholders,
  Badge,
  Button,
  Card,
  EmailTemplateEditor,
  Label,
  MissingConfig,
  Switch,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
  useToast,
} from '@tumaet/prompt-ui-components'
import { MailWarningIcon } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { ResultsReleasedMailSettings } from '../../../../interfaces/resultsReleasedMail'
import { useGetCoursePhaseMetaData } from '../../../hooks/useGetCoursePhaseMetaData'
import { isTemplateComplete, parseResultsReleasedMail } from './utils'

const resultsMailPlaceholders = [
  { placeholder: '{{coursePhaseName}}', description: 'The name of this course phase' },
  { placeholder: '{{coursePhaseLink}}', description: 'The link to this course phase in PROMPT' },
]

export const ResultsReleasedMailCard = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const { toast } = useToast()
  const courseMailingIsConfigured = useGetMailingIsConfigured()
  const { data: coursePhase } = useGetCoursePhaseMetaData()

  const [initialSettings, setInitialSettings] = useState<ResultsReleasedMailSettings | null>(null)
  const [settings, setSettings] = useState<ResultsReleasedMailSettings>(() =>
    parseResultsReleasedMail(undefined),
  )
  const initializedPhaseIdRef = useRef<string | null>(null)

  useEffect(() => {
    if (!coursePhase || initializedPhaseIdRef.current === coursePhase.id) return

    const parsed = parseResultsReleasedMail(coursePhase)
    setSettings(parsed)
    setInitialSettings(parsed)
    initializedPhaseIdRef.current = coursePhase.id
  }, [coursePhase])

  const { mutate: updateCoursePhase, isPending: isSaving } = useModifyCoursePhase(
    () => {
      setInitialSettings(settings)
      toast({ title: 'Results mail settings updated' })
    },
    () => {
      toast({
        title: 'Failed to update results mail settings',
        description: 'Please try again later.',
        variant: 'destructive',
      })
    },
  )

  const isModified = JSON.stringify(initialSettings) !== JSON.stringify(settings)
  const sendOnReleaseAvailable = courseMailingIsConfigured && isTemplateComplete(settings)

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setSettings((prev) => ({ ...prev, [name]: value }))
  }

  const handleSave = () => {
    if (!coursePhase) return

    const mailingSettings =
      (coursePhase.restrictedData?.mailingSettings as Record<string, unknown>) ?? {}
    updateCoursePhase({
      id: coursePhase.id,
      restrictedData: {
        mailingSettings: { ...mailingSettings, resultsReleasedMail: settings },
      },
    })
  }

  return (
    <Card className='border-border shadow-sm'>
      <div className='space-y-6 p-6'>
        <div className='space-y-2'>
          <div className='flex items-center justify-between gap-2'>
            <h2 className='text-xl font-semibold text-foreground'>Results Released Mailing</h2>
            {isModified && (
              <Badge variant='outline' className='border-yellow-300 bg-yellow-100 text-yellow-800'>
                Unsaved Changes
              </Badge>
            )}
          </div>
          <p className='max-w-3xl text-sm leading-6 text-muted-foreground'>
            Notify students by mail when you release the results. Each student receives the mail
            once, also if you release the results again later.
          </p>
        </div>

        <MissingConfig
          elements={
            courseMailingIsConfigured
              ? []
              : [
                  {
                    title: 'Course Sender Information',
                    description:
                      'The course has no `Reply To Email Address` set yet. Please configure it in the course settings.',
                    link: `/management/course/${courseId}/settings`,
                    icon: MailWarningIcon,
                  },
                ]
          }
        />

        <div className='flex items-center justify-between gap-4'>
          <div className='space-y-0.5'>
            <Label htmlFor='sendOnRelease'>Notify Students on Release</Label>
            <p className='text-sm text-muted-foreground'>
              Send the mail below automatically when the results are released.
            </p>
          </div>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div>
                  <Switch
                    id='sendOnRelease'
                    disabled={!sendOnReleaseAvailable}
                    checked={sendOnReleaseAvailable && settings.sendOnRelease}
                    onCheckedChange={(checked) =>
                      setSettings((prev) => ({ ...prev, sendOnRelease: checked }))
                    }
                  />
                </div>
              </TooltipTrigger>
              {!sendOnReleaseAvailable && (
                <TooltipContent>
                  Configure course mailing and add a mail subject and content first.
                </TooltipContent>
              )}
            </Tooltip>
          </TooltipProvider>
        </div>

        <AvailableMailPlaceholders customAdditionalPlaceholders={resultsMailPlaceholders} />

        {/* the tiptap editor reads its content only once, so wait for the stored template */}
        {initialSettings && (
          <EmailTemplateEditor
            subject={settings.subject}
            content={settings.content}
            onInputChange={handleInputChange}
            label='Results Released'
            subjectHTMLLabel='subject'
            contentHTMLLabel='content'
            placeholders={[...availablePlaceholders, ...resultsMailPlaceholders].map(
              ({ placeholder }) => placeholder,
            )}
          />
        )}

        <div className='flex justify-end'>
          <Button onClick={handleSave} disabled={!isModified || isSaving || !coursePhase}>
            {isSaving ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </div>
    </Card>
  )
}
