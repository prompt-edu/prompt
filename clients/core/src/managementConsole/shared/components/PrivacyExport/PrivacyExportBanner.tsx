import { Download, Loader2, ShieldCheck } from 'lucide-react'
import { PrivacyExport } from '@core/network/queries/privacyStudentDataExport'
import { Button } from '@tumaet/prompt-ui-components'

interface PrivacyExportBannerProps {
  inProgress: boolean
  privacyExport: PrivacyExport
}

export function PrivacyExportBanner({ inProgress, privacyExport }: PrivacyExportBannerProps) {
  return (
    <div className='rounded-lg border border-gray-200 bg-gray-50 p-4 flex items-center justify-between'>
      <div className='flex items-center gap-3'>
        <div className='lg:mx-2'>
          {inProgress ? (
            <Loader2 className='animate-spin h-5 w-5 text-gray-500' />
          ) : (
            <ShieldCheck className='h-5 w-5 text-gray-500' />
          )}
        </div>
        <div>
          <p className='font-semibold text-gray-900'>
            {inProgress ? 'Collecting your data…' : 'Export ready'}
          </p>
          <p className='text-xs text-gray-500 mt-0.5'>
            Requested on {new Date(privacyExport.date_created).toLocaleString()}
          </p>
          {!inProgress && (
            <p className='text-xs text-gray-500 mt-0.5'>
              Files available until {new Date(privacyExport.available_until).toLocaleString()}
            </p>
          )}
        </div>
      </div>
      {!inProgress && (
        <Button>
          <Download className='mr-2 h-4 w-4' />
          Download All
        </Button>
      )}
    </div>
  )
}
