import {
  Button,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@tumaet/prompt-ui-components'
import { Download, Upload } from 'lucide-react'
import { type ReactNode, useRef } from 'react'
import { downloadImportTemplate } from '../utils/downloadImportTemplate'

interface ImportUploadStepProps {
  fileName: string | null
  headers: string[]
  rows: Record<string, string>[]
  error: string | null
  onFile: (file: File) => void
}

const PREVIEW_ROW_COUNT = 5

export const ImportUploadStep = ({
  fileName,
  headers,
  rows,
  error,
  onFile,
}: ImportUploadStepProps): ReactNode => {
  const fileInputRef = useRef<HTMLInputElement>(null)

  const previewRows = rows.slice(0, PREVIEW_ROW_COUNT)

  return (
    <div className='space-y-4 min-w-0'>
      <div className='flex items-center justify-between'>
        <p className='text-sm text-muted-foreground'>
          Upload a CSV file with your students. First Name, Last Name, University ID and Email are
          required.
        </p>
        <Button variant='outline' size='sm' onClick={downloadImportTemplate}>
          <Download className='h-4 w-4 mr-2' />
          Template
        </Button>
      </div>

      <div className='flex items-center gap-3'>
        <input
          type='file'
          accept='.csv,text/csv'
          ref={fileInputRef}
          className='hidden'
          onChange={(event) => {
            const file = event.target.files?.[0]
            if (file) {
              onFile(file)
            }
            event.target.value = ''
          }}
        />
        <Button onClick={() => fileInputRef.current?.click()}>
          <Upload className='h-4 w-4 mr-2' />
          Select CSV
        </Button>
        {fileName && <span className='text-sm text-muted-foreground'>{fileName}</span>}
      </div>

      {error && <p className='text-sm text-destructive'>{error}</p>}

      {headers.length > 0 && (
        <div className='min-w-0'>
          <p className='text-sm font-medium mb-2'>
            Preview ({rows.length} row{rows.length === 1 ? '' : 's'})
          </p>
          <div className='w-full max-h-[300px] overflow-auto rounded-md border'>
            <Table>
              <TableHeader>
                <TableRow>
                  {headers.map((header) => (
                    <TableHead key={header} className='min-w-[140px] bg-muted whitespace-nowrap'>
                      {header}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {previewRows.map((row, rowIndex) => (
                  <TableRow key={rowIndex}>
                    {headers.map((header) => (
                      <TableCell key={header} className='whitespace-nowrap'>
                        {row[header]}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      )}
    </div>
  )
}
