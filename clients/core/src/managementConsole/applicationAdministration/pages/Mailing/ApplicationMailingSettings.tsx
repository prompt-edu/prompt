import { useEffect, useState } from 'react'
import { ApplicationMailingMetaData } from '../../interfaces/applicationMailingMetaData'
import { parseApplicationMailingMetaData } from './utils/parseApplicaitonMailingMetaData'
import { useModifyCoursePhase } from '../../hooks/useModifyCoursePhase'
import { UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import {
  applicationMailingPlaceholders,
  CustomApplicationPlaceHolder,
} from './components/CustomApplicationPlaceHolder'
import { EmailTemplateEditor } from '@tumaet/prompt-ui-components'
import {
  Button,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  ManagementPageHeader,
  useToast,
} from '@tumaet/prompt-ui-components'
import { SettingsCard } from './components/SettingsCard'
import { useApplicationStore } from '../../zustand/useApplicationStore'
import { useGetMailingIsConfigured } from '@tumaet/prompt-shared-state'
import { MissingConfig, MissingConfigItem } from '@tumaet/prompt-ui-components'
import { MailWarningIcon } from 'lucide-react'

export const ApplicationMailingSettings = () => {
  const { phaseId, courseId } = useParams<{ courseId: string; phaseId: string }>()
  const { toast } = useToast()
  const { coursePhase } = useApplicationStore()
  const [initialMetaData, setInitialMetaData] = useState<ApplicationMailingMetaData | null>(null)
  const [applicationMailingMetaData, setApplicationMailingMetaData] =
    useState<ApplicationMailingMetaData>({
      confirmationMailSubject: '',
      confirmationMailContent: '',
      failedMailSubject: '',
      failedMailContent: '',
      passedMailSubject: '',
      passedMailContent: '',
      sendConfirmationMail: false,
      sendRejectionMail: false,
      sendAcceptanceMail: false,
    })

  const isModified = JSON.stringify(initialMetaData) !== JSON.stringify(applicationMailingMetaData)

  const courseMailingIsConfigured = useGetMailingIsConfigured()
  const [missingConfigs, setMissingConfigs] = useState<MissingConfigItem[]>([])

  // Updating state
  const { mutate: mutateCoursePhase } = useModifyCoursePhase(
    () => {
      toast({
        title: 'Application mailing settings updated',
      })
    },
    () => {
      toast({
        title: 'Error updating application mailing settings',
        description: 'Please try again later',
        variant: 'destructive',
      })
    },
  )

  useEffect(() => {
    if (coursePhase?.restrictedData) {
      const parsedMetaData = parseApplicationMailingMetaData(coursePhase.restrictedData)
      setApplicationMailingMetaData(parsedMetaData)
      setInitialMetaData(parsedMetaData)
    }
  }, [coursePhase])

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setApplicationMailingMetaData((prev) => ({ ...prev, [name]: value }))
  }

  const handleSwitchChange = (name: string) => {
    setApplicationMailingMetaData((prev) => ({
      ...prev,
      [name]: !prev[name as keyof ApplicationMailingMetaData],
    }))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const updatedCoursePhase: UpdateCoursePhase = {
      id: phaseId ?? '',
      restrictedData: {
        mailingSettings: applicationMailingMetaData,
      },
    }
    mutateCoursePhase(updatedCoursePhase)
  }

  useEffect(() => {
    if (!courseMailingIsConfigured) {
      setMissingConfigs([
        {
          title: 'Course Sender Information',
          description:
            'The course has not yet any set `Reply To Email Address` set. Please make sure to configure this in the course settings.',
          link: `/management/course/${courseId}/settings`,
          icon: MailWarningIcon,
        },
      ])
    }
  }, [courseId, courseMailingIsConfigured])

  return (
    <div className='space-y-6'>
      <ManagementPageHeader>Application Mailing Settings</ManagementPageHeader>
      <MissingConfig elements={missingConfigs} />
      <SettingsCard
        applicationMailingMetaData={applicationMailingMetaData}
        handleSwitchChange={handleSwitchChange}
        isModified={isModified}
      />
      <h2 className='text-2xl font-bold'>Mailing Templates </h2>

      <CustomApplicationPlaceHolder />
      {/* ensures that tiptap editor is only loaded after receiving meta data */}
      {initialMetaData && (
        <Tabs defaultValue='confirmation' className='w-full'>
          <TabsList className='grid w-full grid-cols-3'>
            <TabsTrigger value='confirmation'>1. Confirmation</TabsTrigger>
            <TabsTrigger value='acceptance'>2. Acceptance</TabsTrigger>
            <TabsTrigger value='rejection'>3. Rejection</TabsTrigger>
          </TabsList>
          <TabsContent value='confirmation'>
            <EmailTemplateEditor
              subject={applicationMailingMetaData.confirmationMailSubject}
              content={applicationMailingMetaData.confirmationMailContent}
              onInputChange={handleInputChange}
              label='Confirmation'
              subjectHTMLLabel='confirmationMailSubject'
              contentHTMLLabel='confirmationMailContent'
              placeholders={applicationMailingPlaceholders.map(
                (placeholder) => placeholder.placeholder,
              )}
            />
          </TabsContent>
          <TabsContent value='acceptance'>
            <EmailTemplateEditor
              subject={applicationMailingMetaData.passedMailSubject}
              content={applicationMailingMetaData.passedMailContent}
              onInputChange={handleInputChange}
              label='Acceptance'
              subjectHTMLLabel='passedMailSubject'
              contentHTMLLabel='passedMailContent'
              placeholders={applicationMailingPlaceholders.map(
                (placeholder) => placeholder.placeholder,
              )}
            />
          </TabsContent>
          <TabsContent value='rejection'>
            <EmailTemplateEditor
              subject={applicationMailingMetaData.failedMailSubject}
              content={applicationMailingMetaData.failedMailContent}
              onInputChange={handleInputChange}
              label='Rejection'
              subjectHTMLLabel='failedMailSubject'
              contentHTMLLabel='failedMailContent'
              placeholders={applicationMailingPlaceholders.map(
                (placeholder) => placeholder.placeholder,
              )}
            />
          </TabsContent>
        </Tabs>
      )}

      <div className='justify-end flex'>
        <Button onClick={handleSubmit} type='submit' className='ml-auto' disabled={!isModified}>
          Save Changes
        </Button>
      </div>
    </div>
  )
}
