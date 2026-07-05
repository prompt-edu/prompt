import { postAdditionalScore } from '@core/network/mutations/postAdditionalScore'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Upload } from 'lucide-react'
import { useCallback, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { AdditionalScoreUpload } from '../../../../interfaces/additionalScore/additionalScoreUpload'
import type { IndividualScore } from '../../../../interfaces/additionalScore/individualScore'
import type { ApplicationParticipation } from '../../../../interfaces/applicationParticipation'
import { AssessmentScoreUploadPage1, type Page1Ref } from './components/AssessmentScoreUploadPage1'
import { AssessmentScoreUploadPage2, type Page2Ref } from './components/AssessmentScoreUploadPage2'
import { AssessmentScoreUploadPage3 } from './components/AssessmentScoreUploadPage3'
import { matchStudents } from './utils/matchStudents'

interface AssessmentScoreUploadProps {
  applications: ApplicationParticipation[]
}

export default function AssessmentScoreUpload({ applications }: AssessmentScoreUploadProps) {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const [state, setState] = useState({
    page: 1,
    additionalScores: [] as IndividualScore[],
    unmatchedApplications: [] as ApplicationParticipation[],
    rowsWithError: [] as string[][],
    numberOfBelowThreshold: null as number | null,
    open: false,
  })

  const page1Ref = useRef<Page1Ref>(null)
  const page2Ref = useRef<Page2Ref>(null)
  const { toast } = useToast()

  const onError = (message: string) => {
    toast({
      title: 'Error',
      description: message,
      variant: 'destructive',
    })
  }

  const resetStates = useCallback(() => {
    setState({
      page: 1,
      additionalScores: [],
      unmatchedApplications: [],
      rowsWithError: [],
      numberOfBelowThreshold: null,
      open: false,
    })
    if (page1Ref.current) {
      page1Ref.current.reset()
    }
    if (page2Ref.current) {
      page2Ref.current.reset()
    }
  }, [])

  const { mutate: mutateSendScore } = useMutation({
    mutationFn: (additionalScore: AdditionalScoreUpload) => {
      return postAdditionalScore(phaseId ?? 'undefined', additionalScore)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['application_participations', phaseId],
      })
      queryClient.invalidateQueries({
        queryKey: ['application_participations', 'students', phaseId],
      })
      toast({
        title: 'Successfully added the scores!',
        variant: 'default',
      })
      resetStates()
    },
    onError: () => {
      toast({
        title: 'Failed to upload the scores',
        description: 'An error occurred. Please try again later.',
        variant: 'destructive',
      })
      resetStates()
    },
  })

  const handlePageNavigation = (next: boolean) => {
    if (next) {
      if (state.page === 1 && page1Ref.current?.validate()) {
        setState((prev) => ({ ...prev, page: 2 }))
      } else if (state.page === 2 && page2Ref.current?.validate()) {
        const page1Values = page1Ref.current?.getValues()
        const page2Values = page2Ref.current?.getValues()
        if (page1Values && page2Values) {
          matchStudents(
            page2Values.csvData,
            page2Values.matchBy,
            page2Values.matchColumn,
            page2Values.scoreColumn,
            page1Values.hasThreshold ? parseFloat(page1Values.threshold) : null,
            onError,
            applications,
            setState,
          )
          setState((prev) => ({ ...prev, page: 3 }))
        }
      } else if (state.page === 3) {
        const page1Values = page1Ref.current?.getValues()
        const newScore: AdditionalScoreUpload = {
          name: page1Values?.scoreName ?? '',
          key: page1Values?.scoreName.toLowerCase().replace(/\s+/g, '') ?? '',
          thresholdActive: !!page1Values?.hasThreshold,
          threshold: page1Values?.hasThreshold ? parseFloat(page1Values.threshold) : 0,
          scores: state.additionalScores,
        }
        mutateSendScore(newScore)
      }
    } else {
      setState((prev) => ({ ...prev, page: Math.max(1, prev.page - 1) }))
    }
  }

  return (
    <Dialog
      open={state.open}
      onOpenChange={(newOpen) => {
        if (!newOpen) {
          resetStates()
        }
        setState((prev) => ({ ...prev, open: newOpen }))
      }}
    >
      <DialogTrigger asChild>
        <Button>
          <Upload className='h-4 w-4 mr-2' />
          Upload Scores
        </Button>
      </DialogTrigger>
      <DialogContent className='sm:max-w-[900px] w-[90vw]'>
        <DialogHeader>
          <DialogTitle>Upload Assessment Scores</DialogTitle>
        </DialogHeader>
        <div className='mt-4'>
          <div style={{ display: state.page === 1 ? 'block' : 'none' }}>
            <AssessmentScoreUploadPage1 ref={page1Ref} />
          </div>
          <div style={{ display: state.page === 2 ? 'block' : 'none' }}>
            <AssessmentScoreUploadPage2 ref={page2Ref} />
          </div>
          <div style={{ display: state.page === 3 ? 'block' : 'none' }}>
            <AssessmentScoreUploadPage3
              matchedCount={state.additionalScores.length}
              unmatchedApplications={state.unmatchedApplications}
              belowThreshold={state.numberOfBelowThreshold}
              rowsWithError={state.rowsWithError}
            />
          </div>
        </div>
        <div className='mt-4 flex justify-between'>
          {state.page !== 1 && (
            <Button onClick={() => handlePageNavigation(false)} disabled={state.page === 1}>
              Previous
            </Button>
          )}
          <Button className='ml-auto' onClick={() => handlePageNavigation(true)}>
            {state.page === 3 ? 'Submit' : 'Next'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
