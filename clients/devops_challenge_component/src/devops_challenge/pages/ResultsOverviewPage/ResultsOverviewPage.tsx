import { Loader2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  ColumnFiltersState,
  useReactTable,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  flexRender,
  SortingState,
  VisibilityState,
} from '@tanstack/react-table'

import { getCoursePhaseParticipations } from '@/network/queries/getCoursePhaseParticipations'
import { CoursePhaseParticipationsWithResolution, PassStatus } from '@tumaet/prompt-shared-state'

import { DeveloperWithInfo } from '../../interfaces/DeveloperWithInfo'
import { getAllDeveloperProfiles } from '../../network/queries/getAllDeveloperProfiles'
import { FilterMenu } from './components/FilterMenu'
import { GroupActionsMenu } from './components/GroupActionsMenu'
import { SelectStudentsDialog } from './components/SelectStudentsDialog'
import { useGetParticipationsWithProfiles } from './hooks/useGetParticipationsWithProfiles'
import { columns } from './columns'
import {
  ErrorPage,
  Button,
  ScrollArea,
  ScrollBar,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  ManagementPageHeader,
  useCustomElementWidth,
} from '@tumaet/prompt-ui-components'

export const ResultsOverviewPage = () => {
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [isSelectDialogOpen, setSelectDialogOpen] = useState(false)
  const [selectCount, setSelectCount] = useState(0)
  const tableWidth = useCustomElementWidth('table-view')

  const { phaseId } = useParams<{ phaseId: string }>()
  const {
    data: coursePhaseParticipations,
    isPending: isCoursePhaseParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
  } = useQuery<CoursePhaseParticipationsWithResolution>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
  })
  const {
    data: developerProfiles,
    isPending: isDeveloperProfilesPending,
    isError: isDeveloperProfileError,
    refetch: refetchDeveloperProfiles,
  } = useQuery<DeveloperWithInfo[]>({
    queryKey: ['developerProfiles', phaseId],
    queryFn: () => getAllDeveloperProfiles(phaseId ?? ''),
  })

  const isError = isParticipationsError || isDeveloperProfileError
  const isPending = isCoursePhaseParticipationsPending || isDeveloperProfilesPending

  const handleRefresh = () => {
    refetchCoursePhaseParticipations()
    refetchDeveloperProfiles()
  }

  const participantsWithProfiles = useGetParticipationsWithProfiles(
    coursePhaseParticipations?.participations || [],
    developerProfiles || [],
  )

  const table = useReactTable({
    data: participantsWithProfiles,
    columns: columns.map((col) => ({ ...col, enableSorting: true })),
    state: { sorting, columnFilters, rowSelection, columnVisibility },
    onSortingChange: setSorting,
    onRowSelectionChange: setRowSelection,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  })

  // When the filters change, update the row selection so that only visible rows remain selected
  useEffect(() => {
    // Get the ids of rows that are visible after filtering
    const visibleRowIDs = new Set(table.getFilteredRowModel().rows.map((row) => row.id))
    // Update rowSelection to only keep selections for visible rows
    setRowSelection((prevSelection) => {
      const newSelection = { ...prevSelection }
      Object.keys(newSelection).forEach((id) => {
        if (!visibleRowIDs.has(id)) {
          delete newSelection[id]
        }
      })
      return newSelection
    })
  }, [columnFilters, table])
  const filteredRowsCount = table.getFilteredRowModel().rows.length
  const totalRowsCount = participantsWithProfiles?.length ?? 0
  const studentsPassedChallengeCount = participantsWithProfiles?.filter(
    (p) => p.profile?.hasPassed,
  ).length
  const studentsPassedCount = participantsWithProfiles?.filter(
    (p) => p.participation.passStatus === PassStatus.PASSED,
  ).length

  if (isError) return <ErrorPage onRetry={handleRefresh} />
  if (isPending)
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )

  return (
    <div id='table-view' className='relative flex flex-col'>
      <div className='space-y-6'>
        <ManagementPageHeader>Developer Profile Management</ManagementPageHeader>
        <div className='flex flex-col lg:flex-row lg:justify-between lg:items-end'>
          <div className='text-sm text-muted-foreground lg:w-1/2'>
            Showing {filteredRowsCount} of {totalRowsCount} applications |{' '}
            {studentsPassedChallengeCount} passed challenge | {studentsPassedCount} accepted
          </div>
          <div className='flex space-x-2 mt-4 lg:mt-0'>
            <Button
              onClick={() => setSelectDialogOpen(true)}
              disabled={studentsPassedChallengeCount <= 0}
            >
              Select First ... Passed Students
            </Button>
            <FilterMenu columnFilters={columnFilters} setColumnFilters={setColumnFilters} />
            <GroupActionsMenu
              selectedRows={table.getSelectedRowModel()}
              onClose={() => table.resetRowSelection()}
            />
          </div>
        </div>
        <div className='rounded-md border' style={{ width: `${tableWidth}px` }}>
          <ScrollArea className='h-[calc(100vh-300px)] overflow-x-scroll'>
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead
                        key={header.id}
                        className='cursor-pointer'
                        onClick={header.column.getToggleSortingHandler()}
                      >
                        {flexRender(header.column.columnDef.header, header.getContext())}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow key={row.id}>
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id} className='font-medium'>
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={columns.length} className='h-24 text-center'>
                      No results.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
            <ScrollBar orientation='horizontal' />
          </ScrollArea>
        </div>

        {isSelectDialogOpen && (
          <SelectStudentsDialog
            isOpen={isSelectDialogOpen}
            onClose={() => setSelectDialogOpen(false)}
            selectCount={selectCount}
            setSelectCount={setSelectCount}
            table={table}
            setRowSelection={setRowSelection}
            studentsPassedChallengeCount={studentsPassedChallengeCount}
          />
        )}
      </div>
    </div>
  )
}
