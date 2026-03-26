import { useForm } from 'react-hook-form'
import { useState } from 'react'
import { AlertCircle, Plus } from 'lucide-react'

import {
  Button,
  Input,
  Label,
  Textarea,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'

import { CreateCategoryRequest } from '../../../../../interfaces/category'

import { useCreateCategory } from '../hooks/useCreateCategory'

export const CreateCategoryForm = ({
  assessmentSchemaID,
  onCancel,
}: {
  assessmentSchemaID: string
  onCancel?: () => void
}) => {
  const [error, setError] = useState<string | undefined>(undefined)
  const { register, handleSubmit, reset } = useForm<CreateCategoryRequest>()
  const { mutate, isPending } = useCreateCategory(setError)

  const onSubmit = (data: CreateCategoryRequest) => {
    mutate(
      { ...data, assessmentSchemaID },
      {
        onSuccess: () => {
          reset()
          onCancel?.()
        },
      },
    )
  }

  return (
    <Card className='shadow-sm transition-all hover:shadow-md'>
      <CardHeader className='pb-3'>
        <CardTitle className='text-lg font-medium flex items-center gap-2'>
          <Plus className='h-4 w-4 text-muted-foreground' />
          Create New Category
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className='space-y-4'>
          <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
            <div className='space-y-2'>
              <Label htmlFor='name' className='text-sm font-medium'>
                Category Name
              </Label>
              <Input
                id='name'
                placeholder='Enter category name'
                className='focus-visible:ring-1'
                {...register('name', { required: true })}
              />
            </div>

            <div className='space-y-2'>
              <Label htmlFor='shortName' className='text-sm font-medium'>
                Short Category Name
              </Label>
              <Input
                id='shortName'
                placeholder='Enter short category name'
                className='focus-visible:ring-1'
                {...register('shortName', { required: true })}
              />
            </div>
          </div>

          <div className='space-y-2'>
            <Label htmlFor='description' className='text-sm font-medium'>
              Description
            </Label>
            <Textarea
              id='description'
              placeholder='Enter category description (optional)'
              className='resize-none min-h-[80px] focus-visible:ring-1'
              {...register('description')}
            />
          </div>

          <div className='space-y-2'>
            <Label htmlFor='weight' className='text-sm font-medium'>
              Weight
            </Label>
            <Input
              id='weight'
              type='number'
              placeholder='Enter weight'
              className='focus-visible:ring-1'
              {...register('weight', {
                required: true,
                valueAsNumber: true,
                validate: (value) => value > 0 || 'Weight must be greater than 0',
              })}
            />
          </div>

          {error && (
            <div className='flex items-center gap-2 text-destructive text-sm p-2 bg-destructive/10 rounded-md'>
              <AlertCircle className='h-4 w-4' />
              <p>{error}</p>
            </div>
          )}

          <div className='flex gap-2'>
            <Button type='submit' disabled={isPending} className='flex-1'>
              {isPending ? 'Creating...' : 'Create Category'}
            </Button>
            {onCancel && (
              <Button
                type='button'
                variant='outline'
                onClick={onCancel}
                disabled={isPending}
                className='flex-1'
              >
                Cancel
              </Button>
            )}
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
