import { type BeforeCapture, DragDropContext, Droppable, type DropResult } from '@hello-pangea/dnd'
import {
  Button,
  Card,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Lock, Plus } from 'lucide-react'
import { useState } from 'react'
import { AssessmentType } from '../../../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../../../interfaces/category'
import { useGetAllCategoriesWithCompetencies } from '../../../hooks/useGetAllCategoriesWithCompetencies'
import { useGetCoursePhaseConfig } from '../../../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../../../hooks/useGetEvaluationCategoriesWithCompetencies'
import { useTutorLabel } from '../../../hooks/useTutorLabel'
import { getSchemaSectionContent } from '../../../schemaSectionContent'
import {
  CATEGORY_DROP_TYPE,
  findCompetencyNameConflict,
  moveCategory,
  moveCompetency,
} from '../../utils/schemaOrder'
import { SchemaPrintReport } from '../SchemaPrintReport'
import { CategoryItem } from './components/CategoryItem'
import { CreateCategoryForm } from './components/CreateCategoryForm'
import { DeleteConfirmDialog } from './components/DeleteConfirmDialog'
import { EditCategoryDialog } from './components/EditCategoryDialog'
import { SchemaTemplateButtons } from './components/SchemaTemplateButtons'
import { useUpdateSchemaOrder } from './hooks/useUpdateSchemaOrder'

interface CategoryListProps {
  assessmentSchemaID: string
  assessmentType: AssessmentType
  hasAssessmentData?: boolean
  schemaName?: string
  schemaDescription?: string
}

const LOCK_BADGE_CLASS = [
  'inline-flex items-center gap-1 rounded-full bg-amber-100 px-3 py-1 text-xs font-medium',
  'text-amber-900 dark:bg-amber-900/40 dark:text-amber-200',
].join(' ')

export const CategoryList = ({
  assessmentSchemaID,
  assessmentType,
  hasAssessmentData = false,
  schemaName,
  schemaDescription,
}: CategoryListProps) => {
  const [categoryToEdit, setCategoryToEdit] = useState<CategoryWithCompetencies | undefined>(
    undefined,
  )
  const [categoryToDelete, setCategoryToDelete] = useState<string | undefined>(undefined)
  const [showAddCategoryForm, setShowAddCategoryForm] = useState(false)
  const [isDraggingCategory, setIsDraggingCategory] = useState(false)
  const { toast } = useToast()
  const { mutate: updateSchemaOrder } = useUpdateSchemaOrder(assessmentType)

  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const { data: assessmentCategories } = useGetAllCategoriesWithCompetencies()
  const { data: selfEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.SELF,
    coursePhaseConfig?.selfEvaluationEnabled ?? false,
  )
  const { data: peerEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.PEER,
    coursePhaseConfig?.peerEvaluationEnabled ?? false,
  )
  const { data: tutorEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.TUTOR,
    coursePhaseConfig?.tutorEvaluationEnabled ?? false,
  )

  const categories =
    assessmentType === AssessmentType.SELF
      ? selfEvaluationCategories
      : assessmentType === AssessmentType.PEER
        ? peerEvaluationCategories
        : assessmentType === AssessmentType.TUTOR
          ? tutorEvaluationCategories
          : assessmentCategories

  const content = getSchemaSectionContent(useTutorLabel())[assessmentType]

  // Collapse every category before a category drag measures them, so they swap by their headers
  const handleBeforeCapture = ({ draggableId }: BeforeCapture) => {
    setIsDraggingCategory(categories.some((category) => category.id === draggableId))
  }

  const handleDragEnd = ({ source, destination, type }: DropResult) => {
    setIsDraggingCategory(false)
    if (
      !destination ||
      (source.droppableId === destination.droppableId && source.index === destination.index)
    ) {
      return
    }

    if (type === CATEGORY_DROP_TYPE) {
      updateSchemaOrder(moveCategory(categories, source.index, destination.index))
      return
    }

    const from = { categoryID: source.droppableId, index: source.index }
    const conflictingName = findCompetencyNameConflict(categories, from, destination.droppableId)
    if (conflictingName) {
      toast({
        title: 'Competency not moved',
        description: `The target category already contains a competency named "${conflictingName}".`,
        variant: 'destructive',
      })
      return
    }

    updateSchemaOrder(
      moveCompetency(categories, from, {
        categoryID: destination.droppableId,
        index: destination.index,
      }),
    )
  }

  return (
    <>
      <Card className='overflow-hidden border-border p-6 shadow-xs print:hidden'>
        <div className='space-y-6'>
          <div className='space-y-2'>
            <div className='flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
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
                          Schema changes are disabled because submitted assessment data already
                          exists for this phase.
                        </p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                )}
              </div>

              <SchemaTemplateButtons
                categories={categories}
                assessmentSchemaID={assessmentSchemaID}
                assessmentType={assessmentType}
                schemaName={schemaName}
                schemaDescription={schemaDescription}
                disabled={hasAssessmentData}
              />
            </div>

            <p className='text-sm leading-6 text-muted-foreground'>
              Review the {content.inlineTitle} structure, category weights, competency descriptions,
              and score-level guidance below.
              {!hasAssessmentData &&
                ' Drag categories and competencies to change their order or to move a competency to another category.'}
            </p>
          </div>

          <div className='space-y-6 border-t border-border pt-6'>
            {categories.length === 0 ? (
              <div className='rounded-xl border border-dashed border-border bg-muted/40 p-6 text-sm text-muted-foreground'>
                This schema does not contain any categories yet. Add the first category to start
                defining the rubric.
              </div>
            ) : (
              <DragDropContext onBeforeCapture={handleBeforeCapture} onDragEnd={handleDragEnd}>
                <Droppable droppableId='categories' type={CATEGORY_DROP_TYPE}>
                  {(provided) => (
                    <div ref={provided.innerRef} {...provided.droppableProps}>
                      {categories.map((category, index) => (
                        <CategoryItem
                          key={category.id}
                          category={category}
                          index={index}
                          setCategoryToEdit={setCategoryToEdit}
                          setCategoryToDelete={setCategoryToDelete}
                          assessmentType={assessmentType}
                          disabled={hasAssessmentData}
                          defaultExpanded
                          collapsed={isDraggingCategory}
                        />
                      ))}
                      {provided.placeholder}
                    </div>
                  )}
                </Droppable>
              </DragDropContext>
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

      <SchemaPrintReport
        categories={categories}
        schemaType={content.title}
        schemaName={schemaName}
        schemaDescription={schemaDescription}
      />
    </>
  )
}
