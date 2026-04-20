import { GithubUsernameInput } from './components/GithubUsernameInput'
import { AssessmentPanel } from './components/AssessmentPanel'
import { useDevOpsChallengeStore } from './zustand/useDevOpsChallengeStore'

export const DevOpsChallengePage = () => {
  const { developerProfile } = useDevOpsChallengeStore()

  return (
    <div className='max-w-xl mx-auto p-4'>
      <div className='space-y-2 mb-6'>
        <h1 className='text-2xl font-bold'>DevOps Challenge</h1>
        <p className='text-gray-500'>Complete the tasks to demonstrate your DevOps skills</p>
      </div>

      <div className='space-y-6'>
        {!developerProfile ? <GithubUsernameInput /> : <AssessmentPanel />}
      </div>
    </div>
  )
}
