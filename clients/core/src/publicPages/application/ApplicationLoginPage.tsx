import { useParams, useNavigate } from 'react-router-dom'
import { ApplicationFormWithDetails } from '@core/interfaces/application/applicationFormWithDetails'
import { getApplicationFormWithDetails } from '@core/network/queries/applicationFormWithDetails'
import { useMutation, useQuery } from '@tanstack/react-query'
import { NonAuthenticatedPageWrapper } from '../shared/components/NonAuthenticatedPageWrapper'
import { ApplicationHeader } from './components/ApplicationHeader'
import { LoadingState } from './components/LoadingState'
import { ErrorState } from './components/ErrorState'
import { useState } from 'react'
import { ApplicationLoginCard } from './components/ApplicationLoginCard'
import { ApplicationFormView } from './pages/ApplicationForm/ApplicationFormView'
import { Student } from '@tumaet/prompt-shared-state'
import { CreateApplicationAnswerText } from '@core/interfaces/application/applicationAnswer/text/createApplicationAnswerText'
import { CreateApplicationAnswerMultiSelect } from '@core/interfaces/application/applicationAnswer/multiSelect/createApplicationAnswerMultiSelect'
import { CreateApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/createApplicationAnswerFileUpload'
import { PostApplication } from '@core/interfaces/application/postApplication'
import { postNewApplicationExtern } from '@core/network/mutations/postApplicationExtern'
import { ApplicationSavingDialog } from './components/ApplicationSavingDialog'

export const ApplicationLoginPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const [selectedContinueAsExternal, setSelectedContinueAsExternal] = useState(false)
  const [selectedContinueWithoutLogin, setSelectedContinueWithoutLogin] = useState(false)
  const [showDialog, setShowDialog] = useState<'saving' | 'success' | 'error' | null>(null)
  const [confirmationMailSent, setConfirmationMailSent] = useState<boolean>(false)

  const {
    data: applicationForm,
    isPending,
    isError,
    error,
  } = useQuery<ApplicationFormWithDetails>({
    queryKey: ['applicationForm', phaseId],
    queryFn: () => getApplicationFormWithDetails(phaseId ?? ''),
  })

  const { mutate: mutateSendApplication, error: mutateError } = useMutation({
    mutationFn: (application: PostApplication) => {
      return postNewApplicationExtern(phaseId ?? 'undefined', application)
    },
    onSuccess: (data) => {
      // Assume the response contains a boolean "confirmationMailSent"
      setConfirmationMailSent(data.confirmationMailSent)
      setShowDialog('success')
    },
    onError: () => {
      setShowDialog('error')
    },
  })

  const handleSubmit = (
    student: Student,
    answersText: CreateApplicationAnswerText[],
    answersMultiSelect: CreateApplicationAnswerMultiSelect[],
    answersFileUpload: CreateApplicationAnswerFileUpload[],
  ) => {
    const application: PostApplication = {
      student,
      answersText: answersText,
      answersMultiSelect: answersMultiSelect,
      answersFileUpload: answersFileUpload,
    }
    setShowDialog('saving')
    mutateSendApplication(application)
  }

  const handleCloseDialog = () => {
    setShowDialog(null)
  }

  if (isPending) {
    return (
      <NonAuthenticatedPageWrapper withLoginButton={false}>
        <LoadingState />
      </NonAuthenticatedPageWrapper>
    )
  }

  if (isError || !applicationForm) {
    return (
      <NonAuthenticatedPageWrapper withLoginButton={false}>
        <ErrorState error={error} onBack={() => navigate('/')} />
      </NonAuthenticatedPageWrapper>
    )
  }

  const { applicationPhase } = applicationForm
  const externalStudentsAllowed = applicationPhase.externalStudentsAllowed
  const universityLoginAvailable = applicationPhase.universityLoginAvailable

  const continueWithOutLogin = (isExternalStudent: boolean) => {
    setSelectedContinueAsExternal(isExternalStudent)
    setSelectedContinueWithoutLogin(true)
  }

  return (
    <NonAuthenticatedPageWrapper withLoginButton={false}>
      <div className='max-w-4xl mx-auto space-y-6'>
        <ApplicationHeader applicationPhase={applicationPhase} onBackClick={() => navigate('/')} />
        {!selectedContinueWithoutLogin && (
          <ApplicationLoginCard
            universityLoginAvailable={universityLoginAvailable}
            externalStudentsAllowed={externalStudentsAllowed}
            onContinueWithoutLogin={continueWithOutLogin}
          />
        )}
        {selectedContinueWithoutLogin &&
          (externalStudentsAllowed && selectedContinueAsExternal ? (
            // enforce Login and get MatriculationNumber and university login from token
            <ApplicationFormView
              questionsText={applicationForm.questionsText}
              questionsMultiSelect={applicationForm.questionsMultiSelect}
              questionsFileUpload={applicationForm.questionsFileUpload}
              coursePhaseId={phaseId}
              onSubmit={handleSubmit}
            />
          ) : (
            // continue with form that allows to enter university login and matriculationNumber
            <ApplicationFormView
              questionsText={applicationForm.questionsText}
              questionsMultiSelect={applicationForm.questionsMultiSelect}
              questionsFileUpload={applicationForm.questionsFileUpload}
              coursePhaseId={phaseId}
              onSubmit={handleSubmit}
              allowEditUniversityData={true}
              student={{
                firstName: '',
                lastName: '',
                email: '',
                matriculationNumber: '',
                universityLogin: '',
                hasUniversityAccount: true,
              }}
            />
          ))}
      </div>
      <ApplicationSavingDialog
        showDialog={showDialog}
        onClose={handleCloseDialog}
        onNavigateBack={() => navigate('/')}
        errorMessage={mutateError?.message}
        confirmationMailSent={confirmationMailSent}
      />
    </NonAuthenticatedPageWrapper>
  )
}
