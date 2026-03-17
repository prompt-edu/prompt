import { StudentTable } from '../shared/components/StudentTable/StudentTable'

export const StudentsPage = () => {
  return (
    <div className='flex flex-col gap-6 w-full'>
      <h1 className='text-3xl font-bold tracking-tight'>Students</h1>
      <StudentTable />
    </div>
  )
}
