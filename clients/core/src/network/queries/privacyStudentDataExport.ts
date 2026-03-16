import { axiosInstance } from '@/network/configService'

export const requestStudentDataExport = async (): Promise<[]> => {
  try {
    return (
      await axiosInstance.post('/api/privacy/student-data-export', {
        headers: {
          'Content-Type': 'application/json-path+json',
        },
      })
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
