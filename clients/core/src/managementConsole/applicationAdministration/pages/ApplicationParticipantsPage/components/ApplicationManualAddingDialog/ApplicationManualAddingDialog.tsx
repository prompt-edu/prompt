import { useCallback, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Loader2, Plus, AlertCircle, ArrowLeft } from 'lucide-react'
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  ScrollArea,
  Alert,
  AlertDescription,
  AlertTitle,
  useToast,
} from '@tumaet/prompt-ui-components'
import { ApplicationForm } from '../../../../interfaces/form/applicationForm'
import { getApplicationForm } from '@core/network/queries/applicationForm'
import { UniversitySelection } from './components/UniversitySelection'
import { ApplicationFormView } from '@core/publicPages/application/pages/ApplicationForm/ApplicationFormView'
import { StudentSearch } from './components/StudentSearch'
import { Student } from '@tumaet/prompt-shared-state'
import { postNewApplicationManual } from '@core/network/mutations/postApplicationManual'
import { PostApplication } from '@core/interfaces/application/postApplication'
import { CreateApplicationAnswerText } from '@core/interfaces/application/applicationAnswer/text/createApplicationAnswerText'
import { CreateApplicationAnswerMultiSelect } from '@core/interfaces/application/applicationAnswer/multiSelect/createApplicationAnswerMultiSelect'
import { CreateApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/createApplicationAnswerFileUpload'
import { ApplicationParticipation } from '../../../../interfaces/applicationParticipation'

interface ApplicationManualAddingDialog {
  existingApplications: ApplicationParticipation[]
}

export const ApplicationManualAddingDialog = ({
  existingApplications,
}: ApplicationManualAddingDialog) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [state, setState] = useState({
    page: 1,
    universityAccount: false,
  })
  const [selectedStudent, setSelectedStudent] = useState<Student | null>(null)

  const resetStates = useCallback(() => {
    setState({
      page: 1,
      universityAccount: false,
    })
    setSelectedStudent(null)
  }, [])

  const {
    data: fetchedApplicationForm,
    isPending: isFetchingApplicationForm,
    isError: isApplicationFormError,
    error: fetchingError,
    refetch: refetchApplicationForm,
  } = useQuery<ApplicationForm>({
    queryKey: ['application_form', phaseId],
    queryFn: () => getApplicationForm(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const { mutate: mutateSendApplication } = useMutation({
    mutationFn: (manualApplication: PostApplication) => {
      return postNewApplicationManual(phaseId ?? 'undefined', manualApplication)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['application_participations', 'students', phaseId],
      })
      toast({
        title: 'Application added',
        description: 'The application has been successfully added',
      })
    },
    onError: (error) => {
      toast({
        title: 'Error',
        description: `The application could not be added. ${error.message}`,
        variant: 'destructive',
      })
    },
  })

  const submitManualApplication = (
    student: Student,
    answersText: CreateApplicationAnswerText[],
    answersMultiSelect: CreateApplicationAnswerMultiSelect[],
    answersFileUpload: CreateApplicationAnswerFileUpload[],
  ) => {
    const manualApplication: PostApplication = {
      student: student,
      answersText: answersText,
      answersMultiSelect: answersMultiSelect,
      answersFileUpload: answersFileUpload,
    }
    mutateSendApplication(manualApplication)
    resetStates()
    setDialogOpen(false)
  }

  const renderContent = () => {
    switch (state.page) {
      case 1:
        return (
          <UniversitySelection
            setHasUniversityAccount={(hasUniversityAccount) => {
              setState((prev) => ({ ...prev, universityAccount: hasUniversityAccount, page: 2 }))
            }}
          />
        )
      case 2:
        return state.universityAccount ? (
          <StudentSearch
            onSelect={(student) => {
              setSelectedStudent(student)
              setState((prev) => ({ ...prev, page: 3 }))
            }}
            existingApplications={existingApplications}
          />
        ) : (
          <ScrollArea className='max-h-[calc(90vh-150px)]'>
            <ApplicationFormView
              questionsText={fetchedApplicationForm?.questionsText ?? []}
              questionsMultiSelect={fetchedApplicationForm?.questionsMultiSelect ?? []}
              questionsFileUpload={fetchedApplicationForm?.questionsFileUpload ?? []}
              coursePhaseId={phaseId}
              onSubmit={submitManualApplication}
            />
          </ScrollArea>
        )
      case 3:
        return (
          <ScrollArea className='max-h-[calc(90vh-150px)]'>
            <ApplicationFormView
              questionsText={fetchedApplicationForm?.questionsText ?? []}
              questionsMultiSelect={fetchedApplicationForm?.questionsMultiSelect ?? []}
              questionsFileUpload={fetchedApplicationForm?.questionsFileUpload ?? []}
              coursePhaseId={phaseId}
              student={
                selectedStudent !== null
                  ? selectedStudent
                  : {
                      firstName: '',
                      lastName: '',
                      email: '',
                      hasUniversityAccount: true,
                    }
              }
              allowEditUniversityData={selectedStudent === null}
              onSubmit={submitManualApplication}
            />
          </ScrollArea>
        )
      default:
        return null
    }
  }

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(newOpen) => {
        if (!newOpen) {
          resetStates()
        }
        setDialogOpen(newOpen)
      }}
    >
      <DialogTrigger asChild>
        <Button>
          <Plus className='mr-2 h-4 w-4' />
          Add Application
        </Button>
      </DialogTrigger>
      <DialogContent className='sm:max-w-[900px] w-[90vw] max-h-[90vh]'>
        <DialogHeader>
          <DialogTitle>Add New Application</DialogTitle>
        </DialogHeader>
        {isFetchingApplicationForm ? (
          <div className='flex justify-center items-center h-64'>
            <Loader2 className='h-12 w-12 animate-spin text-primary' />
          </div>
        ) : isApplicationFormError ? (
          <Alert variant='destructive'>
            <AlertCircle className='h-4 w-4' />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>
              Failed to fetch application form. {fetchingError?.message}
              <Button variant='link' onClick={() => refetchApplicationForm()}>
                Try again
              </Button>
            </AlertDescription>
          </Alert>
        ) : (
          renderContent()
        )}
        <DialogFooter className='sm:justify-start'>
          {state.page !== 1 && (
            <Button
              className='mr-auto'
              onClick={() => setState((prev) => ({ ...prev, page: prev.page - 1 }))}
            >
              <ArrowLeft className='mr-2 h-4 w-4' /> Back
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
