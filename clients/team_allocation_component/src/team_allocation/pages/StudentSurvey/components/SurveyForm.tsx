import { useMutation } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { SkillLevel, SkillResponse } from '../../../interfaces/skillResponse'
import { SurveyResponse } from '../../../interfaces/surveyResponse'
import { TeamPreference } from '../../../interfaces/teamPreference'
import { postSurveyResponse } from '../../../network/mutations/postSurveyResponse'
import { SurveyForm } from '../../../interfaces/surveyForm'
import { useParams } from 'react-router-dom'
import { AlertTriangle, CheckCircle, CheckCircle2, Loader2 } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle, Button } from '@tumaet/prompt-ui-components'
import { TeamRanking } from './TeamRanking'
import { SkillRanking } from './SkillRanking'
import dayjs from 'dayjs'
import { SurveyStatusBar } from './SurveyStatusBar'

interface SurveyFormProps {
  surveyForm: SurveyForm
  surveyResponse: SurveyResponse | undefined
  isStudent: boolean
}

export const SurveyFormComponent = ({ surveyForm, surveyResponse, isStudent }: SurveyFormProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const preferenceMode = surveyForm.preferenceMode ?? 'teams'
  const isFieldMode = preferenceMode === 'fields'

  // Local state for ranking teams and rating skills
  const [teamRanking, setTeamRanking] = useState<string[]>([])
  const [skillRatings, setSkillRatings] = useState<Record<string, SkillLevel | undefined>>({})
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [submitSuccess, setSubmitSuccess] = useState<boolean>(false)

  // Store the original data for comparison
  const [initialTeamRanking, setInitialTeamRanking] = useState<string[]>([])
  const [initialSkillRatings, setInitialSkillRatings] = useState<
    Record<string, SkillLevel | undefined>
  >({})

  // When form and response are loaded, initialize the local state.
  useEffect(() => {
    if (!surveyForm) return

    const currentTeamIDs = new Set(surveyForm.teams.map((t) => t.id))

    let newTeamRanking: string[]
    const savedPreferences = isFieldMode
      ? surveyResponse?.fieldPreferences
      : surveyResponse?.teamPreferences

    if (savedPreferences?.length) {
      const sorted = [...savedPreferences].sort((a, b) => a.preference - b.preference)
      const savedTeamIDs = sorted.map((pref) => pref.teamID)
      const savedMatchCurrent =
        savedTeamIDs.length === surveyForm.teams.length &&
        savedTeamIDs.every((id) => currentTeamIDs.has(id))

      if (savedMatchCurrent) {
        newTeamRanking = savedTeamIDs
      } else {
        newTeamRanking = surveyForm.teams.map((team) => team.id).sort(() => Math.random() - 0.5)
      }
    } else {
      // We randomize the order in the beginning to avoid bias
      newTeamRanking = surveyForm.teams.map((team) => team.id).sort(() => Math.random() - 0.5)
    }
    setTeamRanking(newTeamRanking)
    setInitialTeamRanking(newTeamRanking)

    const currentSkillIDs = new Set(surveyForm.skills.map((s) => s.id))
    const newSkillRatings: Record<string, SkillLevel | undefined> = {}
    if (surveyResponse?.skillResponses?.length) {
      surveyResponse.skillResponses.forEach((sr) => {
        if (currentSkillIDs.has(sr.skillID)) {
          newSkillRatings[sr.skillID] = sr.skillLevel
        }
      })
    }

    setSkillRatings(newSkillRatings)
    setInitialSkillRatings(newSkillRatings)
  }, [surveyForm, surveyResponse, isFieldMode])

  // Mutation for submitting the updated survey response.
  const updateSurveyResponseMutation = useMutation({
    mutationFn: (response: SurveyResponse) => postSurveyResponse(phaseId ?? '', response),
    onSuccess: () => {
      setSubmitError(null)
      setSubmitSuccess(true)
      // Reset the initial state to match the newly submitted data
      setInitialTeamRanking(teamRanking)
      setInitialSkillRatings(skillRatings)
    },
    onError: () => {
      setSubmitError('Failed to update your survey response. Please try again.')
      setSubmitSuccess(false)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    // Create team preferences from the ordering.
    const preferences: TeamPreference[] = teamRanking.map((teamID, index) => ({
      teamID,
      preference: index + 1, // rank 1 for the first item, 2 for second, etc.
    }))

    // Create skill responses from the ratings (include only the ones that are defined).
    const skillResponses: SkillResponse[] = Object.entries(skillRatings)
      .filter(([, skillLevel]) => skillLevel !== undefined) // only include selected ratings
      .map(([skillID, skillLevel]) => ({
        skillID,
        skillLevel: skillLevel!, // non-null assertion since we filtered out undefined
      }))

    const response: SurveyResponse = {
      skillResponses,
      preferenceMode,
      ...(isFieldMode ? { fieldPreferences: preferences } : { teamPreferences: preferences }),
    }
    updateSurveyResponseMutation.mutate(response)
  }

  // Compute status for the top of the page
  let status: 'Not submitted' | 'Submitted' | 'Modified' = 'Not submitted'
  if (surveyResponse) {
    // Check if the local states are the same as the initial ones
    const sameTeamRanking = JSON.stringify(teamRanking) === JSON.stringify(initialTeamRanking)
    const sameSkillRatings = JSON.stringify(skillRatings) === JSON.stringify(initialSkillRatings)
    const skillsSubmitted = Object.values(initialSkillRatings).some(
      (rating) => rating !== undefined,
    )
    status = skillsSubmitted
      ? sameTeamRanking && sameSkillRatings
        ? 'Submitted'
        : 'Modified'
      : 'Not submitted'
  }

  const allSkillsSelected = surveyForm.skills.every((skill) => skillRatings[skill.id] !== undefined)

  const now = dayjs()
  const deadlineTime = dayjs(surveyForm.deadline)
  const diff = deadlineTime.diff(now, 'second')
  const hasDeadlinePassed = diff <= 0

  return (
    <div className='space-y-8'>
      {submitSuccess ? (
        <div className='text-center space-y-4'>
          <div className='flex flex-col items-center space-y-2 text-green-500'>
            <CheckCircle className='h-12 w-12' />
            <h2 className='text-2xl font-semibold'>Success</h2>
          </div>
          <p className='text-muted-foreground max-w-md mx-auto'>
            You have successfully submitted the Survey. You can re-submit the survey until the{' '}
            {dayjs(surveyForm.deadline).format('MMM D, YYYY [at] h:mm A')}, if you want to change
            your answers.
          </p>
          <div className='pt-4'>
            <Button onClick={() => setSubmitSuccess(false)}>Go back to survey</Button>
          </div>
        </div>
      ) : (
        <>
          <SurveyStatusBar deadline={surveyForm.deadline} status={status} />

          <form onSubmit={handleSubmit} className='space-y-8'>
            {/* Teams Ranking Section */}
            <TeamRanking
              teamRanking={teamRanking}
              teams={surveyForm.teams}
              setTeamRanking={setTeamRanking}
              disabled={!isStudent || updateSurveyResponseMutation.isPending || hasDeadlinePassed}
              preferenceMode={preferenceMode}
            />

            {/* Skills Rating Section */}
            <SkillRanking
              skills={surveyForm.skills}
              skillRatings={skillRatings}
              setSkillRatings={setSkillRatings}
              disabled={!isStudent || updateSurveyResponseMutation.isPending || hasDeadlinePassed}
            />

            {submitError && (
              <Alert variant='destructive' className='mb-4 w-full'>
                <AlertTriangle className='h-4 w-4' />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{submitError}</AlertDescription>
              </Alert>
            )}

            {submitSuccess && (
              <div className='flex items-center gap-2 text-green-600'>
                <CheckCircle2 className='h-5 w-5' />
                <span>Survey response submitted successfully!</span>
              </div>
            )}

            <Button
              type='submit'
              size='lg'
              disabled={
                !isStudent ||
                updateSurveyResponseMutation.isPending ||
                !allSkillsSelected ||
                hasDeadlinePassed
              }
              className='w-full sm:w-auto'
            >
              {updateSurveyResponseMutation.isPending ? (
                <>
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  Submitting...
                </>
              ) : (
                'Submit Survey Response'
              )}
            </Button>
          </form>
        </>
      )}
    </div>
  )
}
