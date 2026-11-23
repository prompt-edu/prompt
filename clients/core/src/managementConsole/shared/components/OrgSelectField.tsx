import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import type { FieldValues, Path, UseFormReturn } from 'react-hook-form'
import type { NewCourseOrg } from '../hooks/useNewCourseOrg'

interface OrgFormValues extends FieldValues {
  orgID?: string
}

interface OrgSelectFieldProps<T extends OrgFormValues> extends NewCourseOrg {
  form: UseFormReturn<T>
}

const NO_ORG = 'none'
const ORG_FIELD = 'orgID'

// Submits only once a required org is chosen
export const requireOrg =
  <T extends OrgFormValues>(
    form: UseFormReturn<T>,
    required: boolean,
    onSubmit: (data: T) => void,
  ) =>
  (data: T) => {
    if (required && !data.orgID) {
      form.setError(ORG_FIELD as Path<T>, { message: 'Choose an org for the course' })
      return
    }
    onSubmit(data)
  }

export const OrgSelectField = <T extends OrgFormValues>({
  form,
  orgs,
  required,
}: OrgSelectFieldProps<T>) => {
  if (orgs.length === 0) {
    return null
  }
  return (
    <FormField
      control={form.control}
      name={ORG_FIELD as Path<T>}
      render={({ field }) => (
        <FormItem>
          <FormLabel>
            Org {!required && <span className='text-muted-foreground font-normal'>(optional)</span>}
          </FormLabel>
          <Select
            onValueChange={(value) => field.onChange(value === NO_ORG ? undefined : value)}
            value={field.value ?? (required ? '' : NO_ORG)}
          >
            <FormControl>
              <SelectTrigger>
                <SelectValue placeholder='Choose an org' />
              </SelectTrigger>
            </FormControl>
            <SelectContent>
              {!required && <SelectItem value={NO_ORG}>No org</SelectItem>}
              {orgs.map((org) => (
                <SelectItem key={org.id} value={org.id}>
                  {org.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FormDescription>The Admins of the org manage all of its courses.</FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
