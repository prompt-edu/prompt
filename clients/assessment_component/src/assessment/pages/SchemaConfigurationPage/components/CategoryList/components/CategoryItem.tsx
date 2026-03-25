import { useState } from 'react'
import { ChevronRight, ChevronDown, Edit, Trash2, Plus } from 'lucide-react'

import { Button } from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../../../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../../../../interfaces/category'

import { CompetencyItem } from './CompetencyItem/CompetencyItem'
import { CreateCompetencyForm } from './CreateCompetencyForm'

interface CategoryItemProps {
  category: CategoryWithCompetencies
  setCategoryToEdit: (category: CategoryWithCompetencies | undefined) => void
  setCategoryToDelete: (categoryID: string | undefined) => void
  assessmentType: AssessmentType
  disabled?: boolean
  defaultExpanded?: boolean
}

export const CategoryItem = ({
  category,
  setCategoryToEdit,
  setCategoryToDelete,
  assessmentType,
  disabled = false,
  defaultExpanded = false,
}: CategoryItemProps) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded)
  const [showAddCompetencyForm, setShowAddCompetencyForm] = useState(false)

  const toggleExpand = () => {
    setIsExpanded(!isExpanded)
  }

  return (
    <div key={category.id} className='mb-6'>
      <div className='flex items-center mb-4'>
        <button
          onClick={toggleExpand}
          className='p-1 mr-2 rounded-sm hover:bg-muted focus:outline-none'
          aria-expanded={isExpanded}
          aria-controls={`content-${category.id}`}
        >
          {isExpanded ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
        </button>
        <h2 className='text-xl font-semibold tracking-tight flex-grow'>{category.name}</h2>
        <div className='flex items-center gap-2'>
          <Button
            variant='ghost'
            size='icon'
            className='h-7 w-7'
            onClick={() => setCategoryToEdit(category)}
            aria-label={`Edit ${category.name}`}
            disabled={disabled}
          >
            <Edit size={16} />
          </Button>
          <Button
            variant='ghost'
            size='icon'
            className='h-7 w-7'
            onClick={() => setCategoryToDelete(category.id)}
            aria-label={`Delete ${category.name}`}
            disabled={disabled}
          >
            <Trash2 size={16} className='text-destructive' />
          </Button>
        </div>
      </div>

      {isExpanded && (
        <div>
          {category.competencies.length === 0 ? (
            <p className='text-sm text-muted-foreground italic'>No competencies available yet.</p>
          ) : (
            <div className='grid gap-4'>
              {category.competencies.map((competency) => (
                <CompetencyItem
                  key={competency.id}
                  competency={competency}
                  categoryID={category.id}
                  assessmentType={assessmentType}
                  disabled={disabled}
                />
              ))}
            </div>
          )}
          <div className='py-4 border-t mt-2'>
            {showAddCompetencyForm ? (
              <CreateCompetencyForm
                categoryID={category.id}
                onCancel={() => setShowAddCompetencyForm(false)}
              />
            ) : (
              <Button
                variant='outline'
                disabled={disabled}
                className='w-full border-dashed flex items-center justify-center p-4 hover:bg-muted/50 transition-colors'
                onClick={() => setShowAddCompetencyForm(true)}
              >
                <Plus className='h-4 w-4 mr-2 text-muted-foreground' />
                <span className='text-muted-foreground'>Add Competency to {category.name}</span>
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
