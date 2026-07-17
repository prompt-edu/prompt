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
import { Pencil } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import type {
  AssessmentSchema,
  UpdateAssessmentSchemaRequest,
} from '../../../interfaces/assessmentSchema'
import { useUpdateAssessmentSchema } from '../hooks/useUpdateAssessmentSchema'

interface RenameSchemaDialogProps {
  schema: AssessmentSchema
  onError: (error: string | undefined) => void
}

export const RenameSchemaDialog = ({ schema, onError }: RenameSchemaDialogProps) => {
  const [isDialogOpen, setIsDialogOpen] = useState(false)

  const updateSchemaMutation = useUpdateAssessmentSchema(schema.id, onError)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UpdateAssessmentSchemaRequest>({
    defaultValues: { name: schema.name, description: schema.description },
  })

  const openDialog = (open: boolean) => {
    if (open) {
      reset({ name: schema.name, description: schema.description })
      onError(undefined)
    }
    setIsDialogOpen(open)
  }

  const onSubmit = (data: UpdateAssessmentSchemaRequest) => {
    updateSchemaMutation.mutate(data, {
      onSuccess: () => {
        setIsDialogOpen(false)
        onError(undefined)
      },
    })
  }

  const handleCancel = () => {
    setIsDialogOpen(false)
    reset({ name: schema.name, description: schema.description })
    onError(undefined)
  }

  return (
    <Dialog open={isDialogOpen} onOpenChange={openDialog}>
      <DialogTrigger asChild>
        <Button variant='outline' size='icon' aria-label='Rename schema'>
          <Pencil className='h-4 w-4' />
        </Button>
      </DialogTrigger>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Rename Assessment Schema</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className='space-y-4'>
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
              {...register('description')}
              placeholder='Schema description'
              rows={3}
            />
          </div>

          <div className='flex justify-end gap-2'>
            <Button type='button' variant='outline' onClick={handleCancel}>
              Cancel
            </Button>
            <Button type='submit' disabled={updateSchemaMutation.isPending}>
              {updateSchemaMutation.isPending ? 'Saving...' : 'Save'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
