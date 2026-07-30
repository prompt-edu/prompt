import { Button, useToast } from '@tumaet/prompt-ui-components'
import { Download, Upload } from 'lucide-react'
import { useRef } from 'react'

import type { AssessmentType } from '../../../../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../../../../interfaces/category'
import { triggerTextDownload } from '../../../../../network/queries/exportStudentAssessment'
import {
  type AssessmentSchemaTemplate,
  buildAssessmentSchemaTemplate,
  parseAssessmentSchemaTemplate,
} from '../../../utils/assessmentSchemaTemplate'
import { useImportAssessmentSchema } from '../hooks/useImportAssessmentSchema'

interface SchemaTemplateButtonsProps {
  categories: CategoryWithCompetencies[]
  assessmentSchemaID: string
  assessmentType: AssessmentType
  schemaName?: string
  schemaDescription?: string
  disabled?: boolean
}

const toFilenameSlug = (value: string | undefined): string => {
  const slug = (value ?? '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return slug || 'assessment-schema'
}

export const SchemaTemplateButtons = ({
  categories,
  assessmentSchemaID,
  assessmentType,
  schemaName,
  schemaDescription,
  disabled = false,
}: SchemaTemplateButtonsProps) => {
  const { toast } = useToast()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const { mutate: importSchema, isPending: isImporting } = useImportAssessmentSchema(
    assessmentSchemaID,
    assessmentType,
  )

  const handleExport = () => {
    const template = buildAssessmentSchemaTemplate(categories, {
      name: schemaName,
      description: schemaDescription,
    })
    triggerTextDownload(
      JSON.stringify(template, null, 2),
      `${toFilenameSlug(schemaName)}-template.json`,
      'application/json',
    )
  }

  const handleFileSelected = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) {
      return
    }

    let content: string
    try {
      content = await file.text()
    } catch {
      toast({
        title: 'Import failed',
        description: 'The selected file could not be read.',
        variant: 'destructive',
      })
      return
    }

    let template: AssessmentSchemaTemplate
    try {
      template = parseAssessmentSchemaTemplate(content)
    } catch (error) {
      toast({
        title: 'Invalid template file',
        description: error instanceof Error ? error.message : 'The file could not be parsed.',
        variant: 'destructive',
      })
      return
    }

    importSchema(template, {
      onSuccess: (result) => {
        if (result.errors.length > 0) {
          const preview = result.errors.slice(0, 3).join(' ')
          const remaining = result.errors.length - 3
          const suffix = remaining > 0 ? ` (+${remaining} more)` : ''
          toast({
            title: 'Import completed with errors',
            description: `Imported ${result.importedCategories} categories and ${result.importedCompetencies} competencies. ${result.errors.length} item(s) failed: ${preview}${suffix}`,
            variant: 'destructive',
          })
          return
        }
        toast({
          title: 'Template imported',
          description: `Imported ${result.importedCategories} categories and ${result.importedCompetencies} competencies.`,
        })
      },
      onError: () => {
        toast({
          title: 'Import failed',
          description: 'The template could not be imported. Please try again.',
          variant: 'destructive',
        })
      },
    })
  }

  return (
    <div className='flex items-center gap-2'>
      <input
        ref={fileInputRef}
        type='file'
        accept='application/json,.json'
        className='hidden'
        onChange={handleFileSelected}
      />
      <Button variant='outline' size='sm' onClick={handleExport} disabled={categories.length === 0}>
        <Download className='mr-2 h-4 w-4' />
        Export
      </Button>
      <Button
        variant='outline'
        size='sm'
        onClick={() => fileInputRef.current?.click()}
        disabled={disabled || isImporting}
      >
        <Upload className='mr-2 h-4 w-4' />
        {isImporting ? 'Importing...' : 'Import'}
      </Button>
    </div>
  )
}
