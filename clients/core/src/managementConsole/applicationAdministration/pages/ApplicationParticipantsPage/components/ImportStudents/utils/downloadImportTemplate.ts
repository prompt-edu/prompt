const TEMPLATE_HEADERS = [
  'First Name',
  'Last Name',
  'University ID',
  'Email',
  'Matriculation Number',
  'Gender',
  'Nationality',
  'Study Program',
  'Study Degree',
  'Current Semester',
]

const EXAMPLE_ROW = [
  'Jane',
  'Doe',
  'ab12cde',
  'jane.doe@tum.de',
  '01234567',
  'female',
  'DE',
  'Computer Science',
  'bachelor',
  '3',
]

/**
 * Triggers a download of a CSV template containing the required and recognized optional headers
 * plus one example row.
 */
export const downloadImportTemplate = (): void => {
  const content = `${TEMPLATE_HEADERS.join(',')}\n${EXAMPLE_ROW.join(',')}\n`
  const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', 'student-import-template.csv')
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
