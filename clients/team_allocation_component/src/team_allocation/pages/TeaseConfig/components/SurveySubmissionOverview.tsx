import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'
import { ArrowDown, ArrowUp, ArrowUpDown, Check, X } from 'lucide-react'
import { useState } from 'react'
import type { TeaseStudent } from '../../../interfaces/tease/student'
import { StudentDetailDialog } from './StudentDetailDialog'

interface StudentTableProps {
  students: TeaseStudent[]
}

export const SurveySubmissionOverview = ({ students }: StudentTableProps) => {
  const [selectedStudent, setSelectedStudent] = useState<TeaseStudent | null>(null)
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [sortConfig, setSortConfig] = useState<{
    key: 'firstName' | 'lastName' | 'email' | 'surveyCompleted'
    direction: 'ascending' | 'descending'
  } | null>(null)

  const handleStudentClick = (student: TeaseStudent) => {
    setSelectedStudent(student)
    setIsDialogOpen(true)
  }

  const requestSort = (key: 'firstName' | 'lastName' | 'email' | 'surveyCompleted') => {
    let direction: 'ascending' | 'descending' = 'ascending'

    if (sortConfig && sortConfig.key === key && sortConfig.direction === 'ascending') {
      direction = 'descending'
    }

    setSortConfig({ key, direction })
  }

  const getSortedStudents = () => {
    if (!sortConfig) return students

    return [...students].sort((a, b) => {
      if (sortConfig.key === 'surveyCompleted') {
        const aHasSubmitted = a.projectPreferences.length > 0
        const bHasSubmitted = b.projectPreferences.length > 0

        if (aHasSubmitted === bHasSubmitted) return 0

        if (sortConfig.direction === 'ascending') {
          return aHasSubmitted ? -1 : 1
        } else {
          return aHasSubmitted ? 1 : -1
        }
      }

      const aValue = a[sortConfig.key]
      const bValue = b[sortConfig.key]

      if (aValue < bValue) {
        return sortConfig.direction === 'ascending' ? -1 : 1
      }
      if (aValue > bValue) {
        return sortConfig.direction === 'ascending' ? 1 : -1
      }
      return 0
    })
  }

  const sortedStudents = getSortedStudents()

  return (
    <>
      <div className='rounded-md border'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className='cursor-pointer' onClick={() => requestSort('firstName')}>
                <div className='flex items-center'>
                  First Name
                  {sortConfig?.key === 'firstName' ? (
                    sortConfig.direction === 'ascending' ? (
                      <ArrowUp className='ml-2 h-4 w-4' />
                    ) : (
                      <ArrowDown className='ml-2 h-4 w-4' />
                    )
                  ) : (
                    <ArrowUpDown className='ml-2 h-4 w-4' />
                  )}
                </div>
              </TableHead>
              <TableHead className='cursor-pointer' onClick={() => requestSort('lastName')}>
                <div className='flex items-center'>
                  Last Name
                  {sortConfig?.key === 'lastName' ? (
                    sortConfig.direction === 'ascending' ? (
                      <ArrowUp className='ml-2 h-4 w-4' />
                    ) : (
                      <ArrowDown className='ml-2 h-4 w-4' />
                    )
                  ) : (
                    <ArrowUpDown className='ml-2 h-4 w-4' />
                  )}
                </div>
              </TableHead>
              <TableHead className='cursor-pointer' onClick={() => requestSort('email')}>
                <div className='flex items-center'>
                  Email
                  {sortConfig?.key === 'email' ? (
                    sortConfig.direction === 'ascending' ? (
                      <ArrowUp className='ml-2 h-4 w-4' />
                    ) : (
                      <ArrowDown className='ml-2 h-4 w-4' />
                    )
                  ) : (
                    <ArrowUpDown className='ml-2 h-4 w-4' />
                  )}
                </div>
              </TableHead>
              <TableHead className='cursor-pointer' onClick={() => requestSort('surveyCompleted')}>
                <div className='flex items-center'>
                  Survey
                  {sortConfig?.key === 'surveyCompleted' ? (
                    sortConfig.direction === 'ascending' ? (
                      <ArrowUp className='ml-2 h-4 w-4' />
                    ) : (
                      <ArrowDown className='ml-2 h-4 w-4' />
                    )
                  ) : (
                    <ArrowUpDown className='ml-2 h-4 w-4' />
                  )}
                </div>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sortedStudents.map((student) => {
              const hasSubmitted = student.projectPreferences.length > 0

              return (
                <TableRow
                  key={student.id}
                  className='cursor-pointer hover:bg-muted/50'
                  onClick={() => handleStudentClick(student)}
                >
                  <TableCell>{student.firstName}</TableCell>
                  <TableCell>{student.lastName}</TableCell>
                  <TableCell>{student.email}</TableCell>
                  <TableCell>
                    {hasSubmitted ? (
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div className='flex items-center'>
                              <div className='bg-green-100 text-green-700 p-1 rounded-full'>
                                <Check className='h-4 w-4' />
                              </div>
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>Survey submitted</TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    ) : (
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div className='flex items-center'>
                              <div className='bg-red-100 text-red-700 p-1 rounded-full'>
                                <X className='h-4 w-4' />
                              </div>
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>Survey not submitted yet</TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    )}
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </div>

      <StudentDetailDialog
        student={selectedStudent}
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
      />
    </>
  )
}

export default SurveySubmissionOverview
