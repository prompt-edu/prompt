import { Control, FieldValues, Path } from 'react-hook-form'
import {
  Textarea,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
  cn,
} from '@tumaet/prompt-ui-components'
import { ScoreLevel } from '@tumaet/prompt-shared-state'

interface AssessmentTextFieldProps<T extends FieldValues> {
  control: Control<T>
  name: Path<T>
  placeholder: string
  completed?: boolean
  getScoreLevel?: () => ScoreLevel | undefined
  onBlur: () => void
}

const isCommentAndExampleRequired = (scoreLevel?: ScoreLevel): boolean => {
  return (
    scoreLevel === ScoreLevel.VeryBad ||
    scoreLevel === ScoreLevel.Bad ||
    scoreLevel === ScoreLevel.Ok
  )
}

export const AssessmentTextField = <T extends FieldValues>({
  control,
  name,
  placeholder,
  completed = false,
  getScoreLevel,
  onBlur,
}: AssessmentTextFieldProps<T>) => {
  const resolveScoreLevel = () => {
    if (getScoreLevel) return getScoreLevel()
  }

  return (
    <div className='flex flex-col h-full'>
      <FormField
        control={control}
        name={name}
        rules={{
          validate: (value) => {
            if (isCommentAndExampleRequired(resolveScoreLevel())) {
              if (!value || value.trim() === '') {
                return name + ' is required for Strongly Disagree, Disagree, and Neutral scores'
              }
            }
            return true
          },
        }}
        render={({ field, fieldState }) => (
          <FormItem className='flex flex-col grow'>
            <FormControl className='grow'>
              <Textarea
                placeholder={completed ? '' : placeholder}
                className={cn(
                  'resize-none text-xs h-full',
                  fieldState.error && 'border border-destructive focus-visible:ring-destructive',
                  completed && 'bg-gray-100 dark:bg-gray-800 cursor-not-allowed opacity-80',
                )}
                disabled={completed}
                readOnly={completed}
                {...field}
                onBlur={() => {
                  field.onBlur()
                  onBlur()
                }}
              />
            </FormControl>
            {!completed && <FormMessage />}
          </FormItem>
        )}
      />
    </div>
  )
}
