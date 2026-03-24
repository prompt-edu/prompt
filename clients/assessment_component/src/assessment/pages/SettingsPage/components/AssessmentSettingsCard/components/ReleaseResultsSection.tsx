import { Button } from '@tumaet/prompt-ui-components'

import { ReleaseConfirmationDialog } from './ReleaseConfirmationDialog'
import { useReleaseAssessmentResults } from '../hooks/useReleaseAssessmentResults'

interface ReleaseResultsSectionProps {
  isSaving: boolean
}

export const ReleaseResultsSection = ({ isSaving }: ReleaseResultsSectionProps) => {
  const {
    resultsReleased,
    showReleaseDialog,
    setShowReleaseDialog,
    confirmRelease,
    isReleasing,
    releaseError,
    completedAssessments,
    totalAssessments,
    allAssessmentsCompleted,
  } = useReleaseAssessmentResults()

  return (
    <>
      <div className='rounded-xl border border-slate-200 bg-slate-50/60 p-4'>
        {resultsReleased ? (
          <div className='space-y-1'>
            <h3 className='text-sm font-semibold text-emerald-900'>Results released</h3>
            <p className='text-sm leading-6 text-emerald-700'>
              Students in this phase can already view the released assessment results based on the
              visibility settings above.
            </p>
          </div>
        ) : (
          <div className='space-y-4'>
            <div className='space-y-1'>
              <h3 className='text-sm font-semibold text-slate-900'>Release results</h3>
              <p className='text-sm leading-6 text-slate-600'>
                Release final assessment results to students after every assessment is marked as
                final. This action cannot be undone.
              </p>
            </div>

            <Button
              onClick={() => setShowReleaseDialog(true)}
              disabled={isSaving || isReleasing || !allAssessmentsCompleted}
              className='w-full'
            >
              {isReleasing
                ? 'Releasing...'
                : `Release Results (${completedAssessments}/${totalAssessments} final)`}
            </Button>

            {!allAssessmentsCompleted && (
              <p className='text-xs leading-5 text-slate-500'>
                All assessments must be marked as final before results can be released.
              </p>
            )}
          </div>
        )}
      </div>

      <ReleaseConfirmationDialog
        open={showReleaseDialog}
        onOpenChange={setShowReleaseDialog}
        onConfirm={confirmRelease}
        isReleasing={isReleasing}
        releaseError={releaseError}
        completedAssessments={completedAssessments}
        totalAssessments={totalAssessments}
      />
    </>
  )
}
