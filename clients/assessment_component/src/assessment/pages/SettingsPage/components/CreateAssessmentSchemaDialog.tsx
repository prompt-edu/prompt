import { useState } from 'react'
import { useForm } from 'react-hook-form'
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Input,
  Label,
  Textarea,
} from '@tumaet/prompt-ui-components'
import { Plus } from 'lucide-react'

import { useCreateAssessmentSchema } from '../hooks/useCreateAssessmentSchema'
import { CreateAssessmentSchemaRequest } from '../../../interfaces/assessmentSchema'

interface CreateAssessmentSchemaDialogProps {
  onError: (error: string | undefined) => void
  disabled?: boolean
}

export const CreateAssessmentSchemaDialog = ({
  onError,
  disabled = false,
}: CreateAssessmentSchemaDialogProps) => {
  const [isDialogOpen, setIsDialogOpen] = useState(false)

  const createSchemaMutation = useCreateAssessmentSchema(onError)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateAssessmentSchemaRequest>()

  const onSubmitNewSchema = (data: CreateAssessmentSchemaRequest) => {
    createSchemaMutation.mutate(data, {
      onSuccess: () => {
        reset()
        setIsDialogOpen(false)
        onError(undefined)
      },
      onError: (err) => onError(err.message),
    })
  }

  const handleCancel = () => {
    setIsDialogOpen(false)
    reset()
    onError(undefined)
  }

  return (
    <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
      <DialogTrigger asChild>
        <Button variant='outline' size='icon' disabled={disabled}>
          <Plus className='h-4 w-4' />
        </Button>
      </DialogTrigger>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Create New Assessment Schema</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmitNewSchema)} className='space-y-4'>
          <div className='space-y-2'>
            <Label htmlFor='name'>Name</Label>
            <Input
              id='name'
              {...register('name', { required: 'Name is required' })}
              placeholder='Schema name'
            />
            {errors.name && <p className='text-sm text-destructive'>{errors.name.message}</p>}
          </div>

          <div className='space-y-2'>
            <Label htmlFor='description'>Description</Label>
            <Textarea
              id='description'
              {...register('description', { required: 'Description is required' })}
              placeholder='Schema description'
              rows={3}
            />
            {errors.description && (
              <p className='text-sm text-destructive'>{errors.description.message}</p>
            )}
          </div>

          <div className='flex justify-end gap-2'>
            <Button type='button' variant='outline' onClick={handleCancel}>
              Cancel
            </Button>
            <Button type='submit' disabled={createSchemaMutation.isPending}>
              {createSchemaMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
