import { useState } from 'react'
import { Edit, Trash2 } from 'lucide-react'

import { Button, Badge } from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../../../../../interfaces/assessmentType'
import { Competency } from '../../../../../../interfaces/competency'

import { useCoursePhaseConfigStore } from '../../../../../../zustand/useCoursePhaseConfigStore'
import { useSelfEvaluationCategoryStore } from '../../../../../../zustand/useSelfEvaluationCategoryStore'
import { usePeerEvaluationCategoryStore } from '../../../../../../zustand/usePeerEvaluationCategoryStore'

import { ScoreLevelSelector } from '../../../../../components/ScoreLevelSelector'

import { DeleteConfirmDialog } from '../DeleteConfirmDialog'

import { EditCompetencyDialog } from './components/EditCompetencyDialog'
import { EvaluationMapping } from './components/EvaluationMapping'

interface CompetencyItemProps {
  competency: Competency
  categoryID: string
  assessmentType: AssessmentType
  disabled?: boolean
}

export const CompetencyItem = ({
  competency,
  categoryID,
  assessmentType,
  disabled = false,
}: CompetencyItemProps) => {
  const [competencyToEdit, setCompetencyToEdit] = useState<Competency | undefined>(undefined)
  const [competencyToDelete, setCompetencyToDelete] = useState<
    | {
        id: string
        name: string
        categoryID: string
      }
    | undefined
  >(undefined)

  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const isEvaluationEnabled =
    coursePhaseConfig?.selfEvaluationEnabled || coursePhaseConfig?.peerEvaluationEnabled

  const { allSelfEvaluationCompetencies } = useSelfEvaluationCategoryStore()
  const { allPeerEvaluationCompetencies } = usePeerEvaluationCategoryStore()

  const currentSelfMapping = competency.mappedFromCompetencies.find((id) =>
    allSelfEvaluationCompetencies.some((comp) => comp.id === id),
  )
  const currentPeerMapping = competency.mappedFromCompetencies.find((id) =>
    allPeerEvaluationCompetencies.some((comp) => comp.id === id),
  )

  return (
    <div>
      <div className='rounded-md border p-4 space-y-4'>
        <div className='flex justify-between items-center gap-2'>
          <div className='flex flex-wrap items-center gap-2'>
            <h3 className='text-base font-medium'>{competency.name}</h3>
            <Badge className='h-5 px-2 text-xs font-medium bg-slate-100 text-slate-700 border-slate-200 hover:bg-slate-100'>
              Weight: {competency.weight}
            </Badge>
          </div>

          <div className='flex'>
            <Button
              variant='ghost'
              size='icon'
              className='h-7 w-7'
              onClick={() => setCompetencyToEdit(competency)}
              aria-label={`Edit ${competency.name}`}
              disabled={disabled}
            >
              <Edit size={16} />
            </Button>
            <Button
              variant='ghost'
              size='icon'
              className='h-7 w-7'
              onClick={() =>
                setCompetencyToDelete({
                  id: competency.id,
                  name: competency.name,
                  categoryID: categoryID,
                })
              }
              aria-label={`Delete ${competency.name}`}
              disabled={disabled}
            >
              <Trash2 size={16} className='text-destructive' />
            </Button>
          </div>
        </div>

        <div className='text-sm text-muted-foreground'>{competency.description}</div>

        <ScoreLevelSelector
          className='lg:col-span-2 grid grid-cols-1 lg:grid-cols-5 gap-1'
          competency={competency}
          onScoreChange={() => {}}
          completed={false}
        />

        {isEvaluationEnabled && assessmentType === AssessmentType.ASSESSMENT && (
          <div>
            <h3 className='text-sm font-medium mb-2'>
              {coursePhaseConfig?.selfEvaluationEnabled && coursePhaseConfig?.peerEvaluationEnabled
                ? 'Self and Peer Evaluation Mapping'
                : coursePhaseConfig?.selfEvaluationEnabled
                  ? 'Self Evaluation Mapping'
                  : 'Peer Evaluation Mapping'}
            </h3>

            <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
              {coursePhaseConfig?.selfEvaluationEnabled && (
                <EvaluationMapping
                  assessmentType={AssessmentType.SELF}
                  competencies={allSelfEvaluationCompetencies}
                  currentMapping={currentSelfMapping}
                  competency={competency}
                />
              )}

              {coursePhaseConfig?.peerEvaluationEnabled && (
                <EvaluationMapping
                  assessmentType={AssessmentType.PEER}
                  competencies={allPeerEvaluationCompetencies}
                  currentMapping={currentPeerMapping}
                  competency={competency}
                />
              )}
            </div>
          </div>
        )}
      </div>

      <EditCompetencyDialog
        open={!!competencyToEdit}
        onOpenChange={(open) => !open && setCompetencyToEdit(undefined)}
        competency={competencyToEdit}
      />

      {competencyToDelete && (
        <DeleteConfirmDialog
          open={!!competencyToDelete}
          onOpenChange={(open) => !open && setCompetencyToDelete(undefined)}
          title='Delete Competency'
          description='Are you sure you want to delete this competency? This action cannot be undone.'
          itemType='competency'
          itemId={competencyToDelete.id}
          categoryId={competencyToDelete.categoryID}
        />
      )}
    </div>
  )
}
