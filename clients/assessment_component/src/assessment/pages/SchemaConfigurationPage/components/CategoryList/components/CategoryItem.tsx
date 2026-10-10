import { Draggable, Droppable } from '@hello-pangea/dnd'
import { Button, cn } from '@tumaet/prompt-ui-components'
import { ChevronDown, ChevronRight, Edit, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'

import type { AssessmentType } from '../../../../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../../../../interfaces/category'

import { COMPETENCY_DROP_TYPE } from '../../../utils/schemaOrder'
import { CompetencyItem } from './CompetencyItem/CompetencyItem'
import { CreateCompetencyForm } from './CreateCompetencyForm'

interface CategoryItemProps {
  category: CategoryWithCompetencies
  index: number
  setCategoryToEdit: (category: CategoryWithCompetencies | undefined) => void
  setCategoryToDelete: (categoryID: string | undefined) => void
  assessmentType: AssessmentType
  disabled?: boolean
  defaultExpanded?: boolean
  // Hides the competencies while a category is dragged, so categories are compared by their headers
  collapsed?: boolean
}

export const CategoryItem = ({
  category,
  index,
  setCategoryToEdit,
  setCategoryToDelete,
  assessmentType,
  disabled = false,
  defaultExpanded = false,
  collapsed = false,
}: CategoryItemProps) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded)
  const [showAddCompetencyForm, setShowAddCompetencyForm] = useState(false)

  const toggleExpand = () => {
    setIsExpanded(!isExpanded)
  }

  return (
    <Draggable draggableId={category.id} index={index} isDragDisabled={disabled}>
      {(provided, snapshot) => (
        <div ref={provided.innerRef} {...provided.draggableProps} className='pb-6'>
          <div
            className={cn(
              'rounded-lg transition-shadow',
              snapshot.isDragging && 'bg-background p-2 shadow-lg ring-1 ring-border',
            )}
          >
            <div
              {...provided.dragHandleProps}
              className={cn(
                'flex items-center rounded-md',
                !collapsed && 'mb-4',
                !disabled && 'cursor-grab',
                snapshot.isDragging && 'cursor-grabbing',
              )}
            >
              <button
                type='button'
                onClick={toggleExpand}
                className='p-1 mr-2 rounded-xs hover:bg-muted focus:outline-hidden'
                aria-expanded={isExpanded}
                aria-controls={`content-${category.id}`}
              >
                {isExpanded ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
              </button>
              <h2 className='text-xl font-semibold tracking-tight grow'>{category.name}</h2>
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

            {isExpanded && !collapsed && (
              <div id={`content-${category.id}`}>
                <Droppable droppableId={category.id} type={COMPETENCY_DROP_TYPE}>
                  {(dropProvided, dropSnapshot) => (
                    <div
                      ref={dropProvided.innerRef}
                      {...dropProvided.droppableProps}
                      className={cn(
                        'min-h-12 rounded-md transition-colors',
                        dropSnapshot.isDraggingOver && 'bg-muted/50',
                      )}
                    >
                      {category.competencies.length === 0 && !dropSnapshot.isDraggingOver && (
                        <p className='pb-4 text-sm text-muted-foreground italic'>
                          No competencies available yet.
                        </p>
                      )}
                      {category.competencies.map((competency, competencyIndex) => (
                        <CompetencyItem
                          key={competency.id}
                          competency={competency}
                          index={competencyIndex}
                          categoryID={category.id}
                          assessmentType={assessmentType}
                          disabled={disabled}
                        />
                      ))}
                      {dropProvided.placeholder}
                    </div>
                  )}
                </Droppable>
                <div className='py-4 border-t'>
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
                      <span className='text-muted-foreground'>
                        Add Competency to {category.name}
                      </span>
                    </Button>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </Draggable>
  )
}
