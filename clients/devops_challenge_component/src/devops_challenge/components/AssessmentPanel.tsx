import { RepositoryInformation } from './RepositoryInformation'
import { Assessment } from './Assessment'
import { PassCountIndicator } from './PassCountIndicator'

export const AssessmentPanel = () => {
  return (
    <div className='grid gap-6 mt-4'>
      <RepositoryInformation />
      <Assessment />
      <PassCountIndicator />
    </div>
  )
}
