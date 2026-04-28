import { useState } from 'react'
import { Lock, Plus } from 'lucide-react'

import {
  Button,
  Card,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'

import { useCategoryStore } from '../../../../zustand/useCategoryStore'
import { useSelfEvaluationCategoryStore } from '../../../../zustand/useSelfEvaluationCategoryStore'
import { usePeerEvaluationCategoryStore } from '../../../../zustand/usePeerEvaluationCategoryStore'
import { useTutorEvaluationCategoryStore } from '../../../../zustand/useTutorEvaluationCategoryStore'

import { AssessmentType } from '../../../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../../../interfaces/category'
import { schemaSectionContent } from '../../../schemaSectionContent'

import { CategoryItem } from './components/CategoryItem'
import { EditCategoryDialog } from './components/EditCategoryDialog'
import { DeleteConfirmDialog } from './components/DeleteConfirmDialog'
import { CreateCategoryForm } from './components/CreateCategoryForm'

interface CategoryListProps {
  assessmentSchemaID: string
  assessmentType: AssessmentType
  hasAssessmentData?: boolean
}

const LOCK_BADGE_CLASS = [
  'inline-flex items-center gap-1 rounded-full bg-amber-100 px-3 py-1 text-xs font-medium',
  'text-amber-900 dark:bg-amber-900/40 dark:text-amber-200',
].join(' ')

export const CategoryList = ({
  assessmentSchemaID,
  assessmentType,
  hasAssessmentData = false,
}: CategoryListProps) => {
  const [categoryToEdit, setCategoryToEdit] = useState<CategoryWithCompetencies | undefined>(
    undefined,
  )
  const [categoryToDelete, setCategoryToDelete] = useState<string | undefined>(undefined)
  const [showAddCategoryForm, setShowAddCategoryForm] = useState(false)

  const { categories: assessmentCategories } = useCategoryStore()
  const { selfEvaluationCategories } = useSelfEvaluationCategoryStore()
  const { peerEvaluationCategories } = usePeerEvaluationCategoryStore()
  const { tutorEvaluationCategories } = useTutorEvaluationCategoryStore()

  const categories =
    assessmentType === AssessmentType.SELF
      ? selfEvaluationCategories
      : assessmentType === AssessmentType.PEER
        ? peerEvaluationCategories
        : assessmentType === AssessmentType.TUTOR
          ? tutorEvaluationCategories
          : assessmentCategories

  const content = schemaSectionContent[assessmentType]

  return (
    <Card className='overflow-hidden border-border p-6 shadow-xs'>
      <div className='space-y-6'>
        <div className='space-y-2'>
          <div className='flex flex-wrap items-center gap-2'>
            <h2 className='text-xl font-semibold tracking-tight text-foreground'>
              Categories and competencies
            </h2>
            {hasAssessmentData && (
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span className={LOCK_BADGE_CLASS}>
                      <Lock className='h-3.5 w-3.5' />
                      Locked by submitted data
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className='max-w-xs'>
                      Schema changes are disabled because submitted assessment data already exists
                      for this phase.
                    </p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            )}
          </div>

          <p className='text-sm leading-6 text-muted-foreground'>
            Review the {content.title.toLowerCase()} structure, category weights, competency
            descriptions, and score-level guidance below.
          </p>
        </div>

        <div className='space-y-6 border-t border-border pt-6'>
          {categories.length === 0 ? (
            <div className='rounded-xl border border-dashed border-border bg-muted/40 p-6 text-sm text-muted-foreground'>
              This schema does not contain any categories yet. Add the first category to start
              defining the rubric.
            </div>
          ) : (
            categories.map((category) => (
              <CategoryItem
                key={category.id}
                category={category}
                setCategoryToEdit={setCategoryToEdit}
                setCategoryToDelete={setCategoryToDelete}
                assessmentType={assessmentType}
                disabled={hasAssessmentData}
                defaultExpanded
              />
            ))
          )}

          {showAddCategoryForm ? (
            <CreateCategoryForm
              assessmentSchemaID={assessmentSchemaID}
              onCancel={() => setShowAddCategoryForm(false)}
            />
          ) : (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div>
                    <Button
                      variant='outline'
                      className='w-full border-dashed p-6'
                      onClick={() => setShowAddCategoryForm(true)}
                      disabled={hasAssessmentData}
                    >
                      <Plus className='mr-2 h-5 w-5 text-muted-foreground' />
                      <span className='text-muted-foreground'>Add category</span>
                    </Button>
                  </div>
                </TooltipTrigger>
                {hasAssessmentData && (
                  <TooltipContent>
                    <p>Cannot add categories when assessment data exists.</p>
                  </TooltipContent>
                )}
              </Tooltip>
            </TooltipProvider>
          )}

          <EditCategoryDialog
            open={!!categoryToEdit}
            onOpenChange={(open) => !open && setCategoryToEdit(undefined)}
            category={categoryToEdit}
            assessmentSchemaID={assessmentSchemaID}
          />

          {categoryToDelete && (
            <DeleteConfirmDialog
              open={!!categoryToDelete}
              onOpenChange={(open) => !open && setCategoryToDelete(undefined)}
              title='Delete Category'
              description={
                'Are you sure you want to delete this category? This action cannot be undone and will delete all competencies within this category.'
              }
              itemType='category'
              itemId={categoryToDelete}
            />
          )}
        </div>
      </div>
    </Card>
  )
}
