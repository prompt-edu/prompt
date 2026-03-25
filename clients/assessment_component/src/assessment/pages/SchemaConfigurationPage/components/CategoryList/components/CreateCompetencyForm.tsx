import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { AlertCircle } from 'lucide-react'

import {
  Button,
  Input,
  Label,
  Textarea,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Alert,
  AlertDescription,
} from '@tumaet/prompt-ui-components'

import type { CreateCompetencyRequest } from '../../../../../interfaces/competency'

import { useCreateCompetency } from '../hooks/useCreateCompetency'

export const CreateCompetencyForm = ({
  categoryID,
  onCancel,
}: {
  categoryID: string
  onCancel?: () => void
}) => {
  const [error, setError] = useState<string | undefined>(undefined)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateCompetencyRequest>({
    defaultValues: { categoryID, weight: 1 },
  })

  const { mutate, isPending } = useCreateCompetency(setError)

  const onSubmit = (data: CreateCompetencyRequest) => {
    mutate(data, {
      onSuccess: () => {
        reset()
        onCancel?.()
      },
    })
  }

  return (
    <Card className='w-full shadow-sm'>
      <CardHeader>
        <CardTitle>Create New Competency</CardTitle>
      </CardHeader>

      <CardContent>
        <form id='competency-form' onSubmit={handleSubmit(onSubmit)} className='space-y-4'>
          <input type='hidden' {...register('categoryID', { required: true })} />

          <div className='grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4 mb-6'>
            <div className='md:col-span-2 space-y-2'>
              <Label htmlFor='name' className='font-medium'>
                Name
              </Label>
              <Input
                id='name'
                placeholder='Enter competency name'
                className={errors.name ? 'border-red-500' : ''}
                aria-invalid={errors.name ? 'true' : 'false'}
                {...register('name', { required: true })}
              />
              {errors.name && <p className='text-sm text-red-500 mt-1'>Name is required</p>}
            </div>

            <div className='space-y-2'>
              <Label htmlFor='shortName' className='font-medium'>
                Short Name
              </Label>
              <Input
                id='shortName'
                placeholder='Enter short competency name'
                className={errors.shortName ? 'border-red-500' : ''}
                aria-invalid={errors.shortName ? 'true' : 'false'}
                {...register('shortName', { required: true })}
              />
              {errors.shortName && (
                <p className='text-sm text-red-500 mt-1'>Short Name is required</p>
              )}
            </div>

            <div className='space-y-2'>
              <Label htmlFor='weight' className='font-medium'>
                Weight
              </Label>
              <Input
                id='weight'
                placeholder='Enter competency weight'
                type='number'
                className={errors.weight ? 'border-red-500' : ''}
                aria-invalid={errors.weight ? 'true' : 'false'}
                {...register('weight', { required: true, valueAsNumber: true })}
              />
              {errors.weight && <p className='text-sm text-red-500 mt-1'>Weight is required</p>}
            </div>

            <div className='sm:col-span-2 md:col-span-4 space-y-2'>
              <Label htmlFor='description' className='font-medium'>
                Description
              </Label>
              <Textarea
                id='description'
                placeholder='Brief description of this competency'
                className={`resize-none h-10 ${errors.description ? 'border-red-500' : ''}`}
                aria-invalid={errors.description ? 'true' : 'false'}
                {...register('description', { required: true })}
              />
              {errors.description && (
                <p className='text-sm text-red-500 mt-1'>Description is required</p>
              )}
            </div>
          </div>

          <div className='space-y-2'>
            <h3 className='text-sm font-medium text-muted-foreground'>Proficiency Levels</h3>

            <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4'>
              <div className='space-y-2'>
                <Label htmlFor='descriptionVeryBad' className='font-medium'>
                  Very Bad
                </Label>
                <Input
                  id='descriptionVeryBad'
                  placeholder='Very bad level description'
                  className={errors.descriptionVeryBad ? 'border-red-500' : ''}
                  aria-invalid={errors.descriptionVeryBad ? 'true' : 'false'}
                  {...register('descriptionVeryBad', { required: true })}
                />
                {errors.descriptionVeryBad && (
                  <p className='text-sm text-red-500 mt-1'>Very bad description is required</p>
                )}
              </div>

              <div className='space-y-2'>
                <Label htmlFor='descriptionBad' className='font-medium'>
                  Bad
                </Label>
                <Input
                  id='descriptionBad'
                  placeholder='Bad level description'
                  className={errors.descriptionBad ? 'border-red-500' : ''}
                  aria-invalid={errors.descriptionBad ? 'true' : 'false'}
                  {...register('descriptionBad', { required: true })}
                />
                {errors.descriptionBad && (
                  <p className='text-sm text-red-500 mt-1'>Bad description is required</p>
                )}
              </div>

              <div className='space-y-2'>
                <Label htmlFor='descriptionOk' className='font-medium'>
                  OK
                </Label>
                <Input
                  id='descriptionOk'
                  placeholder='OK level description'
                  className={errors.descriptionOk ? 'border-red-500' : ''}
                  aria-invalid={errors.descriptionOk ? 'true' : 'false'}
                  {...register('descriptionOk', { required: true })}
                />
                {errors.descriptionOk && (
                  <p className='text-sm text-red-500 mt-1'>OK description is required</p>
                )}
              </div>

              <div className='space-y-2'>
                <Label htmlFor='descriptionGood' className='font-medium'>
                  Good
                </Label>
                <Input
                  id='descriptionGood'
                  placeholder='Good level description'
                  className={errors.descriptionGood ? 'border-red-500' : ''}
                  aria-invalid={errors.descriptionGood ? 'true' : 'false'}
                  {...register('descriptionGood', { required: true })}
                />
                {errors.descriptionGood && (
                  <p className='text-sm text-red-500 mt-1'>Good description is required</p>
                )}
              </div>

              <div className='space-y-2'>
                <Label htmlFor='descriptionVeryGood' className='font-medium'>
                  Very Good
                </Label>
                <Input
                  id='descriptionVeryGood'
                  placeholder='Very good level description'
                  className={errors.descriptionVeryGood ? 'border-red-500' : ''}
                  aria-invalid={errors.descriptionVeryGood ? 'true' : 'false'}
                  {...register('descriptionVeryGood', { required: true })}
                />
                {errors.descriptionVeryGood && (
                  <p className='text-sm text-red-500 mt-1'>Very good description is required</p>
                )}
              </div>
            </div>
          </div>

          <div className='flex gap-2'>
            <Button type='submit' disabled={isPending} className='flex-1'>
              {isPending ? 'Creating...' : 'Create'}
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

        {error && (
          <Alert variant='destructive' className='mt-4'>
            <AlertCircle className='h-4 w-4' />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
      </CardContent>
    </Card>
  )
}
