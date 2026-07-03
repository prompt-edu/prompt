import { Button, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { ChevronLeft, X } from 'lucide-react'
import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useMatchingStore } from '../../zustand/useMatchingStore'
import MatchingResults from './components/MatchingResults'
import { RankingOptions } from './components/RankingOptions'
import { useStudentMatching } from './hooks/useStudentMatching'

export const DataExportPage = () => {
  const path = useLocation().pathname
  const navigate = useNavigate()
  const { uploadedData } = useMatchingStore()
  const { matchedByMatriculation, matchedByName, unmatchedApplications, unmatchedStudents } =
    useStudentMatching()

  const [useScoreAsRank, setUseScoreAsRank] = useState<boolean>(true)

  const handleRankingChange = (useScore: boolean) => {
    setUseScoreAsRank(useScore)
  }

  if (uploadedData?.length === 0) {
    return (
      <div className='flex flex-col items-center justify-center h-[calc(100vh-4rem)]'>
        <p className='text-2xl font-semibold'>No data uploaded</p>
        <Button onClick={() => navigate(path.replace('/export', ''))} className='mt-4'>
          <ChevronLeft />
          Go back
        </Button>
      </div>
    )
  }

  return (
    <div>
      <ManagementPageHeader>Data Export</ManagementPageHeader>
      <RankingOptions onRankingChange={handleRankingChange} useScoreAsRank={true} />

      <MatchingResults
        matchedByMatriculation={matchedByMatriculation}
        matchedByName={matchedByName}
        unmatchedApplications={unmatchedApplications}
        unmatchedStudents={unmatchedStudents}
        useScoreAsRank={useScoreAsRank}
      />
      <Button onClick={() => navigate(-1)} className='mt-4'>
        <X />
        Close
      </Button>
    </div>
  )
}
