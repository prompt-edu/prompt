import { DeleteConfirmation } from '@tumaet/prompt-ui-components'
import { useCallback, useState } from 'react'
import type { CoursePhaseWithPosition } from '../interfaces/coursePhaseWithPosition'

interface DeleteConfirmationReturn {
  deleteDialogIsOpen: boolean
  toBeDeletedComponent: string
  onBeforeDelete: (params: {
    nodes: any[] // TODO fix
    edges: any[]
  }) => Promise<false | { nodes: any[]; edges: any[] }>
  DeleteConfirmationComponent
}

interface UseDeleteConfirmationProps {
  coursePhases: CoursePhaseWithPosition[]
  setIsModified: (modified: boolean) => void
}

export function useDeleteConfirmation({
  coursePhases,
  setIsModified,
}: UseDeleteConfirmationProps): DeleteConfirmationReturn {
  const [deleteDialogIsOpen, setDeleteDialogOpen] = useState(false)
  const [toBeDeletedComponent, setToBeDeletedComponent] = useState('')
  const [deleteConfirmationResolver, setDeleteConfirmationResolver] =
    useState<(value: boolean) => void>()

  const onBeforeDelete = useCallback(
    async ({ nodes: toBeDeletedNodes, edges: toBeDeletedEdges }) => {
      const hasInitialPhase = toBeDeletedNodes.some(
        (node) =>
          node.id && coursePhases.some((phase) => phase.id === node.id && phase.isInitialPhase),
      )
      if (hasInitialPhase) {
        console.log('Cannot delete initial phase')
        return false
      }

      setDeleteDialogOpen(true)
      setToBeDeletedComponent(toBeDeletedNodes.length > 0 ? 'a Course Phase' : 'an Edge')
      const userDecision = await new Promise<boolean>((resolve) => {
        setDeleteConfirmationResolver(() => resolve)
      })
      if (userDecision) {
        setIsModified(true)
        return { nodes: toBeDeletedNodes, edges: toBeDeletedEdges }
      } else {
        return false
      }
    },
    [coursePhases, setIsModified],
  )

  const DeleteConfirmationComponent = (
    <DeleteConfirmation
      isOpen={deleteDialogIsOpen}
      setOpen={setDeleteDialogOpen}
      deleteMessage={`Are you sure you want to delete ${toBeDeletedComponent}?`}
      onClick={(value: boolean) => {
        if (deleteConfirmationResolver) {
          deleteConfirmationResolver(value)
          setDeleteConfirmationResolver(undefined)
        }
      }}
    />
  )

  return {
    deleteDialogIsOpen,
    toBeDeletedComponent,
    onBeforeDelete,
    DeleteConfirmationComponent,
  }
}
