import { useState } from 'react'
import { AlertCircle } from 'lucide-react'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Alert,
  AlertDescription,
} from '@tumaet/prompt-ui-components'

import { useDeleteCategory } from '../hooks/useDeleteCategory'
import { useDeleteCompetency } from '../hooks/useDeleteCompetency'

interface DeleteConfirmDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  itemType: 'category' | 'competency'
  itemId: string
  categoryId?: string
}

export function DeleteConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  itemType,
  itemId,
}: DeleteConfirmDialogProps) {
  const [error, setError] = useState<string | undefined>(undefined)
  const { mutate: deleteCategory, isPending: isDeletingCategory } = useDeleteCategory(setError)
  const { mutate: deleteCompetency, isPending: isDeletingCompetency } =
    useDeleteCompetency(setError)

  const isDeleting = isDeletingCategory || isDeletingCompetency

  const handleDelete = () => {
    if (itemType === 'category') {
      deleteCategory(itemId, {
        onSuccess: () => {
          onOpenChange(false)
        },
      })
    } else if (itemType === 'competency') {
      deleteCompetency(itemId, {
        onSuccess: () => {
          onOpenChange(false)
        },
      })
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>

        {error && (
          <Alert variant='destructive'>
            <AlertCircle className='h-4 w-4' />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={isDeleting}
            className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
          >
            {isDeleting ? 'Deleting...' : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
