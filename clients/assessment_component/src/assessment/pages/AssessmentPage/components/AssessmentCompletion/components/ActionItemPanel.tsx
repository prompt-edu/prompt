import { useAuthStore } from '@tumaet/prompt-shared-state'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  ErrorPage,
} from '@tumaet/prompt-ui-components'
import { AlertCircle, Loader2, Plus } from 'lucide-react'
import { useState } from 'react'
import { useParams } from 'react-router-dom'

import type { ActionItem, UpdateActionItemRequest } from '../../../../../interfaces/actionItem'
import { useStudentAssessmentStore } from '../../../../../zustand/useStudentAssessmentStore'
import { ItemRow } from '../../../../components/ItemRow'
import { useGetActionItemsForStudent } from '../../../../hooks/useGetActionItemsForStudent'
import { useGetCoursePhaseConfig } from '../../../../hooks/useGetCoursePhaseConfig'
import { useCreateActionItem } from '../hooks/useCreateActionItem'
import { useDeleteActionItem } from '../hooks/useDeleteActionItem'
import { useUpdateActionItem } from '../hooks/useUpdateActionItem'
import { DeleteActionItemDialog } from './DeleteActionItemDialog'

interface ActionItemPanelProps {
  readOnly?: boolean
  actionItems?: ActionItem[]
}

export function ActionItemPanel({ readOnly = false, actionItems }: ActionItemPanelProps) {
  const { phaseId, courseParticipationID } = useParams<{
    phaseId: string
    courseParticipationID: string
  }>()
  const [error, setError] = useState<string | undefined>(undefined)
  const [savingItemId, setSavingItemId] = useState<string | undefined>(undefined)
  const [itemValues, setItemValues] = useState<Record<string, string>>({})
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [itemToDelete, setItemToDelete] = useState<string | undefined>(undefined)

  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const { assessmentCompletion } = useStudentAssessmentStore()

  const completed = readOnly || assessmentCompletion?.completed

  const {
    actionItems: fetchedActionItems,
    isPending: isGetActionItemsPending,
    isError,
    refetch,
  } = useGetActionItemsForStudent(!readOnly)

  const { mutate: createActionItem, isPending: isCreatePending } = useCreateActionItem(setError)
  const { mutate: updateActionItem, isPending: isUpdatePending } = useUpdateActionItem(setError)
  const { mutate: deleteActionItem, isPending: isDeletePending } = useDeleteActionItem(setError)

  const { user } = useAuthStore()
  const userName = user ? `${user.firstName} ${user.lastName}` : 'Unknown User'

  const resolvedActionItems = readOnly ? (actionItems ?? []) : fetchedActionItems

  const handleAddActionItem = async () => {
    if (completed) return

    await createActionItem({
      coursePhaseID: phaseId ?? '',
      courseParticipationID: courseParticipationID ?? '',
      action: '',
      author: userName,
    })
  }

  const handleTextChange = (itemId: string, value: string) => {
    setItemValues((prev) => ({ ...prev, [itemId]: value }))
  }

  const handleTextBlur = (itemId: string) => {
    if (completed) return

    const item = resolvedActionItems.find((it) => it.id === itemId)
    const value = itemValues[itemId]

    if (item && value !== undefined && value.trim() !== item.action.trim()) {
      setSavingItemId(item.id)

      const updateRequest: UpdateActionItemRequest = {
        id: item.id,
        coursePhaseID: phaseId ?? '',
        courseParticipationID: courseParticipationID ?? '',
        action: value.trim(),
        author: userName,
      }

      updateActionItem(updateRequest, {
        onSuccess: () => {
          setSavingItemId(undefined)
        },
        onError: () => {
          setSavingItemId(undefined)
        },
      })
    }
  }

  const openDeleteDialog = (itemId: string) => {
    if (!completed) {
      setItemToDelete(itemId)
      setDeleteDialogOpen(true)
    }
  }

  const confirmDelete = () => {
    if (itemToDelete) {
      deleteActionItem(itemToDelete, {
        onSuccess: () => {
          setDeleteDialogOpen(false)
        },
      })
    }
  }

  const cancelDelete = () => {
    setDeleteDialogOpen(false)
    setItemToDelete(undefined)
  }

  const isPending = isGetActionItemsPending || isCreatePending || isUpdatePending || isDeletePending

  if (isError) {
    return <ErrorPage message='Error loading assessments' onRetry={refetch} />
  }

  if (isGetActionItemsPending && !readOnly) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Action Items</CardTitle>
          {coursePhaseConfig?.actionItemsVisible && !readOnly && (
            <p className='text-xs text-gray-500 dark:text-gray-400 mt-1'>
              These action items will be visible to the student once results are released.
            </p>
          )}
        </CardHeader>
        <CardContent className='space-y-2'>
          {readOnly && resolvedActionItems.length === 0 && (
            <div className='text-sm text-muted-foreground'>No action items available.</div>
          )}
          {resolvedActionItems.map((item) => (
            <ItemRow
              key={item.id}
              type='action'
              item={item}
              value={itemValues[item.id] ?? item.action}
              onTextChange={handleTextChange}
              onTextBlur={handleTextBlur}
              onDelete={openDeleteDialog}
              isSaving={savingItemId === item.id}
              isPending={isPending}
              isDisabled={completed || readOnly}
            />
          ))}

          {!completed && (
            <Button
              variant='outline'
              className='w-full border-dashed flex items-center justify-center p-6 hover:bg-muted/50 transition-colors'
              onClick={handleAddActionItem}
              disabled={isPending || completed}
              title={
                completed
                  ? 'Assessment completed - cannot add new action items'
                  : 'Add new action item'
              }
            >
              {isCreatePending ? (
                <Loader2 className='h-5 w-5 mr-2 animate-spin text-muted-foreground' />
              ) : (
                <Plus className='h-5 w-5 mr-2 text-muted-foreground' />
              )}
              <span className='text-muted-foreground'>Add Action Item</span>
            </Button>
          )}

          {error && (
            <div className='flex items-center gap-2 text-destructive text-xs p-2 mt-2 bg-destructive/10 rounded-md'>
              <AlertCircle className='h-3 w-3' />
              <p>{error}</p>
            </div>
          )}
        </CardContent>
      </Card>

      <DeleteActionItemDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        onConfirm={confirmDelete}
        onCancel={cancelDelete}
        isDeleting={isDeletePending}
      />
    </>
  )
}
