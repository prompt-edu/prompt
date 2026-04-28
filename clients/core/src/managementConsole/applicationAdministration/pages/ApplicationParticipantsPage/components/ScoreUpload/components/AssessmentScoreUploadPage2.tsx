import React, { useState, forwardRef, useImperativeHandle } from 'react'
import { Equal } from 'lucide-react'
import {
  Input,
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@tumaet/prompt-ui-components'
import { translations } from '@tumaet/prompt-shared-state'

export interface Page2Ref {
  validate: () => boolean
  getValues: () => {
    file: File | null
    csvData: string[][]
    matchBy: 'email' | 'universityLogin' | 'matriculationNumber'
    matchColumn: string
    scoreColumn: string
  }
  reset: () => void
}

export const AssessmentScoreUploadPage2 = forwardRef<Page2Ref>(
  function AssessmentScoreUploadPage2(props, ref) {
    const [file, setFile] = useState<File | null>(null)
    const [csvData, setCsvData] = useState<string[][]>([])
    const [matchBy, setMatchBy] = useState<'email' | 'universityLogin' | 'matriculationNumber'>(
      'universityLogin',
    )
    const [matchColumn, setMatchColumn] = useState('')
    const [scoreColumn, setScoreColumn] = useState('')
    const [errors, setErrors] = useState<{ [key: string]: string }>({})

    const reset = () => {
      setFile(null)
      setCsvData([])
      setMatchBy('universityLogin')
      setMatchColumn('')
      setScoreColumn('')
      setErrors({})
    }

    const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
      const uploadFile = event.target.files?.[0]
      if (uploadFile) {
        setFile(uploadFile)
        const reader = new FileReader()
        reader.onload = (e) => {
          const text = e.target?.result as string
          const rows = text
            .replace(/\r/g, '') // Remove all \r characters
            .split('\n')
            .map((row) => row.split(';').map((value) => value.replace(/"/g, '')))
          setCsvData(rows)
        }
        reader.readAsText(uploadFile)
      }
    }

    const validate = () => {
      const newErrors: { [key: string]: string } = {}

      if (!file) {
        newErrors.file = 'CSV file is required'
      }

      if (!matchColumn) {
        newErrors.matchColumn = 'Match column is required'
      }

      if (!scoreColumn) {
        newErrors.scoreColumn = 'Score column is required'
      }

      setErrors(newErrors)
      return Object.keys(newErrors).length === 0
    }

    useImperativeHandle(ref, () => ({
      validate,
      getValues: () => ({ file, csvData, matchBy, matchColumn, scoreColumn }),
      reset,
    }))

    return (
      <div className='space-y-6'>
        <div className='space-y-4'>
          <Label htmlFor='csvUpload'>Upload CSV file</Label>
          <div className='flex items-center space-x-2'>
            <Input id='csvUpload' type='file' accept='.csv' onChange={handleFileUpload} />
          </div>
          {file && <p className='text-sm text-muted-foreground'>File uploaded: {file.name}</p>}
          {errors.file && <p className='text-sm text-red-500'>{errors.file}</p>}
        </div>

        {csvData.length > 0 && (
          <>
            <div className='space-y-4'>
              <div className='flex items-end space-x-4'>
                <div className='flex-1 space-y-2'>
                  <Label htmlFor='matchBy'>Match students by</Label>
                  <Select
                    value={matchBy}
                    onValueChange={(value: 'email' | 'universityLogin' | 'matriculationNumber') =>
                      setMatchBy(value)
                    }
                  >
                    <SelectTrigger id='matchBy'>
                      <SelectValue placeholder='Select matching criteria' />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value='universityLogin'>
                        {translations.university['login-name']}
                      </SelectItem>
                      <SelectItem value='email'>Email</SelectItem>
                      <SelectItem value='matriculationNumber'>Matriculation Number</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className='flex items-center pb-2'>
                  <Equal className='h-6 w-6 text-muted-foreground' />
                </div>

                <div className='flex-1 space-y-2'>
                  <Label htmlFor='matchColumn'>Select column to match by</Label>
                  <Select value={matchColumn} onValueChange={setMatchColumn}>
                    <SelectTrigger id='matchColumn'>
                      <SelectValue placeholder='Select a column' />
                    </SelectTrigger>
                    <SelectContent>
                      {csvData?.[0]?.map((header, index) => (
                        <SelectItem key={index} value={header}>
                          {header}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {errors.matchColumn && (
                    <p className='text-sm text-red-500'>{errors.matchColumn}</p>
                  )}
                </div>
              </div>
            </div>
            <div className='space-y-4'>
              <Label htmlFor='scoreColumn'>Select column for scores</Label>
              <Select value={scoreColumn} onValueChange={setScoreColumn}>
                <SelectTrigger id='scoreColumn'>
                  <SelectValue placeholder='Select a column' />
                </SelectTrigger>
                <SelectContent>
                  {csvData[0]?.map((header, index) => (
                    <SelectItem key={index} value={header}>
                      {header}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.scoreColumn && <p className='text-sm text-red-500'>{errors.scoreColumn}</p>}
            </div>

            <div className='mt-4 h-[300px] sm:max-w-[850px] w-[85vw] overflow-hidden flex flex-col'>
              <h4 className='text-sm font-medium mb-2'>CSV Preview</h4>
              <div className='overflow-x-auto overflow-y-auto grow'>
                <Table>
                  <TableHeader>
                    <TableRow>
                      {csvData[0].map((header, index) => (
                        <TableHead key={index} className='min-w-[150px] bg-muted'>
                          {header}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {csvData.slice(1).map((row, rowIndex) => (
                      <TableRow key={rowIndex}>
                        {row.map((cell, cellIndex) => (
                          <TableCell key={cellIndex} className='wrap-break-word'>
                            {cell}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          </>
        )}
      </div>
    )
  },
)
