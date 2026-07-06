import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Download } from 'lucide-react'
import { useState } from 'react'
import { useParams } from 'react-router-dom'

import {
  type AssessmentExportFormat,
  exportStudentAssessment,
  triggerTextDownload,
} from '../../../network/queries/exportStudentAssessment'

interface ExportType {
  label: string
  format: AssessmentExportFormat
  filenameExtension: string
  mimeType: string
  serialize: (data: unknown) => string
}

const exportTypes: ExportType[] = [
  {
    label: 'JSON',
    format: 'json',
    filenameExtension: 'json',
    mimeType: 'application/json',
    serialize: (data) => JSON.stringify(data, null, 2),
  },
]

export const AssessmentExportMenu = () => {
  const { phaseId, courseParticipationID } = useParams<{
    phaseId: string
    courseParticipationID: string
  }>()
  const { toast } = useToast()
  const [isExporting, setIsExporting] = useState(false)

  const handleExport = async (exportType: ExportType) => {
    if (!phaseId || !courseParticipationID) return

    setIsExporting(true)
    try {
      const exportData = await exportStudentAssessment(
        phaseId,
        courseParticipationID,
        exportType.format,
      )
      triggerTextDownload(
        exportType.serialize(exportData),
        `assessment_${courseParticipationID}.${exportType.filenameExtension}`,
        exportType.mimeType,
      )
    } catch {
      toast({
        title: 'Export failed',
        description: 'The assessment could not be exported. Please try again.',
        variant: 'destructive',
      })
    } finally {
      setIsExporting(false)
    }
  }

  return (
    <div className='flex justify-end pt-4'>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant='outline' disabled={isExporting}>
            <Download className='h-4 w-4' />
            Export
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end'>
          <DropdownMenuItem onSelect={() => setTimeout(() => window.print(), 0)}>
            PDF / Print
          </DropdownMenuItem>
          {exportTypes.map((exportType) => (
            <DropdownMenuItem key={exportType.format} onClick={() => handleExport(exportType)}>
              {exportType.label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
