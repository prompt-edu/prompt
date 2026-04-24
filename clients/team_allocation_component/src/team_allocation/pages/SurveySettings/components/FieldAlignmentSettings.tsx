import {
  Badge,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import { Rows3 } from 'lucide-react'

import type { CompanyAllocationConfig } from '../../../interfaces/companyImport'

interface FieldAlignmentSettingsProps {
  coursePhase: CoursePhaseWithMetaData
}

export const FieldAlignmentSettings = ({ coursePhase }: FieldAlignmentSettingsProps) => {
  const companyConfig = coursePhase.restrictedData?.companyAllocationConfig as
    | CompanyAllocationConfig
    | undefined
  const fields = companyConfig?.fields ?? []

  return (
    <Card className='w-full shadow-sm'>
      <CardHeader className='pb-3'>
        <CardTitle className='text-2xl font-bold flex items-center gap-2'>
          <Rows3 className='h-5 w-5' />
          Field Alignment
        </CardTitle>
        <CardDescription className='mt-1.5'>
          Students will rank fields of business for this profile. Company projects remain hidden for
          TEASE.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {fields.length > 0 ? (
          <div className='flex flex-wrap gap-2'>
            {fields.map((field) => (
              <Badge key={field} variant='secondary'>
                {field}
              </Badge>
            ))}
          </div>
        ) : (
          <p className='text-sm text-muted-foreground'>
            Upload and save a company CSV to derive field alignment options.
          </p>
        )}
      </CardContent>
    </Card>
  )
}
