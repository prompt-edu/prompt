import { Badge, Button } from '@tumaet/prompt-ui-components'
import { Edit, Trash2 } from 'lucide-react'
import { useState } from 'react'

import type { AssessmentType } from '../../../../../../interfaces/assessmentType'
import type { Competency } from '../../../../../../interfaces/competency'

import { ScoreLevelSelector } from '../../../../../components/ScoreLevelSelector'

import { DeleteConfirmDialog } from '../DeleteConfirmDialog'

import { EditCompetencyDialog } from './components/EditCompetencyDialog'

interface CompetencyItemProps {
  competency: Competency
  categoryID: string
  assessmentType: AssessmentType
  disabled?: boolean
}

export const CompetencyItem = ({
  competency,
  categoryID,
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

  return (
    <div>
      <div className='rounded-md border p-4 space-y-4'>
        <div className='flex justify-between items-center gap-2'>
          <div className='flex flex-wrap items-center gap-2'>
            <h3 className='text-base font-medium'>{competency.name}</h3>
            <Badge className='h-5 border-border bg-muted px-2 text-xs font-medium text-muted-foreground hover:bg-muted'>
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
          className='grid grid-cols-1 gap-1 md:grid-cols-5'
          competency={competency}
          onScoreChange={() => {}}
          completed={false}
        />
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
