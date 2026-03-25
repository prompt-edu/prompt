import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { AlertCircle } from 'lucide-react'

import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'

import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Textarea,
  Alert,
  AlertDescription,
} from '@tumaet/prompt-ui-components'

import type {
  Competency,
  UpdateCompetencyRequest,
} from '../../../../../../../interfaces/competency'
import { useUpdateCompetency } from '../hooks/useUpdateCompetency'

const updateCompetencySchema = z.object({
  id: z.string(),
  categoryID: z.string().optional(),
  name: z.string().min(1, 'Name is required').optional(),
  shortName: z.string().min(1, 'Short name is required').optional(),
  description: z.string().min(1, 'Description is required').optional(),
  descriptionVeryBad: z.string().min(1, 'Very bad level description is required').optional(),
  descriptionBad: z.string().min(1, 'Bad level description is required').optional(),
  descriptionOk: z.string().min(1, 'OK level description is required').optional(),
  descriptionGood: z.string().min(1, 'Good level description is required').optional(),
  descriptionVeryGood: z.string().min(1, 'Very good level description is required').optional(),
  weight: z
    .number()
    .min(0, 'Weight must be positive')
    .max(100, 'Weight cannot exceed 100')
    .optional(),
})

interface EditCompetencyDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  competency?: Competency
}

export function EditCompetencyDialog({
  open,
  onOpenChange,
  competency,
}: EditCompetencyDialogProps) {
  const [error, setError] = useState<string | undefined>(undefined)
  const { mutate, isPending: isUpdating } = useUpdateCompetency(setError)

  const form = useForm<UpdateCompetencyRequest>({
    defaultValues: {
      id: competency?.id,
      categoryID: competency?.categoryID,
      name: competency?.name,
      shortName: competency?.shortName,
      description: competency?.description,
      descriptionVeryBad: competency?.descriptionVeryBad,
      descriptionBad: competency?.descriptionBad,
      descriptionOk: competency?.descriptionOk,
      descriptionGood: competency?.descriptionGood,
      descriptionVeryGood: competency?.descriptionVeryGood,
      weight: competency?.weight,
    },
    resolver: zodResolver(updateCompetencySchema),
  })

  const handleClose = () => {
    onOpenChange(false)
    setError(undefined)
    form.reset()
  }

  const handleDialogOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      handleClose()
      return
    }

    onOpenChange(nextOpen)
  }

  const onSubmit = (data: UpdateCompetencyRequest) => {
    mutate(data, {
      onSuccess: () => handleClose(),
    })
  }

  useEffect(() => {
    if (competency) {
      form.reset({
        id: competency.id,
        categoryID: competency.categoryID,
        name: competency.name,
        shortName: competency.shortName,
        description: competency.description,
        descriptionVeryBad: competency.descriptionVeryBad,
        descriptionBad: competency.descriptionBad,
        descriptionOk: competency.descriptionOk,
        descriptionGood: competency.descriptionGood,
        descriptionVeryGood: competency.descriptionVeryGood,
        weight: competency.weight,
      })
    }
  }, [competency, form])

  return (
    <Dialog open={open} onOpenChange={handleDialogOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Competency</DialogTitle>
          <DialogDescription>Update the competency details below.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Competency Name</FormLabel>
                  <FormControl>
                    <Input placeholder='Enter competency name' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='shortName'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Short Competency Name</FormLabel>
                  <FormControl>
                    <Input placeholder='Enter short competency name' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='description'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Enter competency description'
                      className='resize-none min-h-[80px]'
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
              <FormField
                control={form.control}
                name='descriptionVeryBad'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Very Bad Level</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='Describe very bad level'
                        className='resize-none min-h-[80px]'
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='descriptionBad'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Bad Level</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='Describe bad level'
                        className='resize-none min-h-[80px]'
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='descriptionOk'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>OK Level</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='Describe OK level'
                        className='resize-none min-h-[80px]'
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='descriptionGood'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Good Level</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='Describe good level'
                        className='resize-none min-h-[80px]'
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='descriptionVeryGood'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Very Good Level</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='Describe very good level'
                        className='resize-none min-h-[80px]'
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='weight'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Weight</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      {...field}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {error && (
              <Alert variant='destructive'>
                <AlertCircle className='h-4 w-4' />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <DialogFooter>
              <Button type='button' variant='outline' onClick={handleClose} disabled={isUpdating}>
                Cancel
              </Button>
              <Button type='submit' disabled={isUpdating}>
                {isUpdating ? 'Saving...' : 'Save'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
