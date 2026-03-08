import { useMutation, UseMutationResult } from '@tanstack/react-query'
import { uploadFile, FileUploadParams, FileResponse } from '@/network/mutations/uploadFile'
import { useState } from 'react'

interface UseFileUploadOptions {
  onSuccess?: (data: FileResponse) => void
  onError?: (error: Error) => void
}

interface UseFileUploadReturn {
  upload: UseMutationResult<FileResponse, Error, FileUploadParams, unknown>
  uploadProgress: number
  isUploading: boolean
}

export const useFileUpload = (options?: UseFileUploadOptions): UseFileUploadReturn => {
  const [uploadProgress, setUploadProgress] = useState(0)

  const upload = useMutation<FileResponse, Error, FileUploadParams>({
    mutationFn: async (params: FileUploadParams) => {
      setUploadProgress(0)
      return uploadFile({
        ...params,
        onUploadProgress: (progressEvent) => {
          const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          setUploadProgress(percentCompleted)
          params.onUploadProgress?.(progressEvent)
        },
      })
    },
    onSuccess: (data) => {
      setUploadProgress(100)
      options?.onSuccess?.(data)
    },
    onError: (error) => {
      setUploadProgress(0)
      options?.onError?.(error)
    },
  })

  return {
    upload,
    uploadProgress,
    isUploading: upload.isPending,
  }
}
