import {
  Button,
  Card,
  CardContent,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
} from '@tumaet/prompt-ui-components'
import { GripVertical, MinusIcon, PlusIcon } from 'lucide-react'
import { DragDropContext, Draggable, Droppable } from '@hello-pangea/dnd'
import { useFieldArray, UseFormReturn } from 'react-hook-form'
import { QuestionConfigFormDataMultiSelect } from '@core/validations/questionConfig'
import { motion } from 'framer-motion'

export function MultiSelectConfig({
  form,
}: {
  form: UseFormReturn<QuestionConfigFormDataMultiSelect>
}) {
  // Destructure the move function from useFieldArray for reordering
  const {
    fields: options,
    append,
    remove,
    move,
  } = useFieldArray({
    control: form.control as any,
    name: 'options',
  })

  const minSelect = form.watch('minSelect') || 0
  const maxSelect = form.watch('maxSelect') || options.length

  // Handle the drag-and-drop reordering logic
  const handleDragEnd = (result: any) => {
    if (!result.destination) return
    // Reorder the items in the array
    move(result.source.index, result.destination.index)
  }

  return (
    <div className='space-y-4'>
      <DragDropContext onDragEnd={handleDragEnd}>
        <div>
          <FormItem>
            <FormLabel>
              Options <span className='text-destructive'> *</span>
            </FormLabel>
            <FormControl>
              <Droppable droppableId='options'>
                {(provided) => (
                  <div ref={provided.innerRef} {...provided.droppableProps} className='space-y-2'>
                    {options.map((option, index) => (
                      <Draggable key={option.id} draggableId={option.id} index={index}>
                        {(prov, snapshot) => (
                          <motion.div
                            ref={prov.innerRef}
                            {...prov.draggableProps}
                            initial={{ opacity: 1 }}
                            animate={{ opacity: snapshot.isDragging ? 0.8 : 1 }}
                            transition={{ duration: 0.2 }}
                          >
                            <Card className='overflow-hidden'>
                              <CardContent className='p-0'>
                                <div className='flex items-center space-x-2 p-2'>
                                  <div
                                    {...prov.dragHandleProps}
                                    className='cursor-move p-2 hover:bg-muted rounded transition-colors'
                                  >
                                    <GripVertical className='h-5 w-5 text-muted-foreground' />
                                  </div>
                                  <FormField
                                    control={form.control}
                                    name={`options.${index}`}
                                    render={({ field }) => (
                                      <Input
                                        {...field}
                                        placeholder={`Option ${index + 1}`}
                                        className='flex-grow'
                                      />
                                    )}
                                  />
                                  <Button
                                    type='button'
                                    variant='ghost'
                                    size='icon'
                                    onClick={() => remove(index)}
                                    disabled={options.length === 1}
                                    className='hover:bg-destructive hover:text-destructive-foreground transition-colors'
                                  >
                                    <MinusIcon className='h-4 w-4' />
                                  </Button>
                                </div>
                              </CardContent>
                            </Card>
                          </motion.div>
                        )}
                      </Draggable>
                    ))}
                    {provided.placeholder}
                  </div>
                )}
              </Droppable>
            </FormControl>
            {/* Array-Level Error Message */}
            {form.formState.errors.options &&
              Array.isArray(form.formState.errors.options) &&
              form.formState.errors.options.length > 0 &&
              form.formState.errors.options[0]?.message && (
                <FormMessage className='text-red-500'>
                  {form.formState.errors.options[0].message}
                </FormMessage>
              )}
            {form.formState.errors.options &&
              !Array.isArray(form.formState.errors.options) &&
              typeof form.formState.errors.options === 'object' &&
              'message' in form.formState.errors.options &&
              typeof form.formState.errors.options.message === 'string' && (
                <FormMessage className='text-red-500'>
                  {form.formState.errors.options.message}
                </FormMessage>
              )}
          </FormItem>
          <Button type='button' variant='outline' className='mt-2' onClick={() => append('')}>
            <PlusIcon className='h-4 w-4 mr-2' />
            Add Option
          </Button>
        </div>
      </DragDropContext>

      <div className='flex space-x-4'>
        <FormField
          control={form.control}
          name='minSelect'
          render={({ field }) => (
            <FormItem className='flex-1'>
              <FormLabel>
                Min Required <span className='text-destructive'> *</span>
              </FormLabel>
              <FormControl>
                <Input
                  type='number'
                  {...field}
                  disabled={options.length === 0}
                  min={0}
                  max={Math.min(maxSelect, options.length)}
                  onChange={(e) => {
                    const value = parseInt(e.target.value)
                    field.onChange(value > maxSelect ? maxSelect : value)
                  }}
                />
              </FormControl>
              <FormDescription>The min amount of options required to be selected.</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='maxSelect'
          render={({ field }) => (
            <FormItem className='flex-1'>
              <FormLabel>
                Max Allowed <span className='text-destructive'> *</span>
              </FormLabel>
              <FormControl>
                <Input
                  type='number'
                  {...field}
                  disabled={options.length === 0}
                  min={Math.max(1, minSelect)}
                  max={options.length}
                  onChange={(e) => {
                    const value = parseInt(e.target.value)
                    field.onChange(value < minSelect ? minSelect : value)
                  }}
                />
              </FormControl>
              <FormDescription>The max amount of options allowed to be selected.</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>
    </div>
  )
}
