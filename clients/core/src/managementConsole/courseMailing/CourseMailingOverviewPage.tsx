import {
  Button,
  DeleteConfirmation,
  ErrorPage,
  LoadingPage,
  ManagementPageHeader,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Copy, Pencil, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { MailCampaignStatusBadge } from './components/MailCampaignStatusBadge'
import {
  useCopyMailCampaign,
  useDeleteMailCampaign,
  useGetMailCampaigns,
  useResendFailedMailCampaign,
} from './hooks/useCourseMailingCampaigns'
import { type MailCampaign, MailCampaignStatus } from './interfaces/mailCampaign'

const formatDate = (value: string | null): string => {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

export const CourseMailingOverviewPage = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()

  const { data: campaigns, isLoading, isError, refetch } = useGetMailCampaigns(courseId ?? '')
  const { mutate: deleteCampaign } = useDeleteMailCampaign(courseId ?? '')
  const { mutate: copyCampaign } = useCopyMailCampaign(courseId ?? '')
  const { mutate: resendFailed } = useResendFailedMailCampaign(courseId ?? '')

  const [campaignToDelete, setCampaignToDelete] = useState<MailCampaign | null>(null)

  if (isLoading) return <LoadingPage />
  if (isError || !campaigns) {
    return <ErrorPage message='Failed to load mail campaigns' onRetry={refetch} />
  }

  const handleDelete = (confirmed: boolean) => {
    if (confirmed && campaignToDelete) {
      deleteCampaign(campaignToDelete.id, {
        onSuccess: () => toast({ title: 'Campaign deleted' }),
        onError: () => toast({ title: 'Failed to delete campaign', variant: 'destructive' }),
      })
    }
    setCampaignToDelete(null)
  }

  const handleCopy = (campaign: MailCampaign) => {
    copyCampaign(campaign.id, {
      onSuccess: () => toast({ title: 'Campaign copied' }),
      onError: () => toast({ title: 'Failed to copy campaign', variant: 'destructive' }),
    })
  }

  const handleResend = (campaign: MailCampaign) => {
    resendFailed(campaign.id, {
      onSuccess: (data) =>
        toast({ title: `Resending to ${data.recipientCount} failed recipients` }),
      onError: () => toast({ title: 'Failed to resend campaign', variant: 'destructive' }),
    })
  }

  return (
    <div>
      <div className='flex items-center justify-between'>
        <ManagementPageHeader>Mailing</ManagementPageHeader>
        <Button onClick={() => navigate(`/management/course/${courseId}/mailing/new`)}>
          <Plus className='mr-2 h-4 w-4' />
          New mail
        </Button>
      </div>

      {campaigns.length === 0 ? (
        <p className='text-muted-foreground'>
          No mail campaigns yet. Create one to send an email to your students.
        </p>
      ) : (
        <div className='rounded-md border'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Recipients</TableHead>
                <TableHead>Created by</TableHead>
                <TableHead>Last changed</TableHead>
                <TableHead>Sent</TableHead>
                <TableHead className='text-right'>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {campaigns.map((campaign) => (
                <TableRow
                  key={campaign.id}
                  className='cursor-pointer'
                  onClick={() => navigate(`/management/course/${courseId}/mailing/${campaign.id}`)}
                >
                  <TableCell className='font-medium'>{campaign.name}</TableCell>
                  <TableCell>
                    <MailCampaignStatusBadge status={campaign.status} />
                  </TableCell>
                  <TableCell>
                    {campaign.sentCount}/{campaign.recipientCount} sent
                    {campaign.failedCount > 0 && (
                      <span className='text-red-500'> ({campaign.failedCount} failed)</span>
                    )}
                  </TableCell>
                  <TableCell>{campaign.createdBy.name || campaign.createdBy.email}</TableCell>
                  <TableCell>{formatDate(campaign.updatedAt)}</TableCell>
                  <TableCell>{formatDate(campaign.sentAt)}</TableCell>
                  <TableCell className='text-right' onClick={(e) => e.stopPropagation()}>
                    <div className='flex justify-end gap-1'>
                      <Button
                        variant='ghost'
                        size='icon'
                        title='Edit'
                        onClick={() =>
                          navigate(`/management/course/${courseId}/mailing/${campaign.id}`)
                        }
                      >
                        <Pencil className='h-4 w-4' />
                      </Button>
                      {campaign.failedCount > 0 && (
                        <Button
                          variant='ghost'
                          size='icon'
                          title='Resend to failed'
                          onClick={() => handleResend(campaign)}
                        >
                          <RefreshCw className='h-4 w-4' />
                        </Button>
                      )}
                      <Button
                        variant='ghost'
                        size='icon'
                        title='Copy'
                        onClick={() => handleCopy(campaign)}
                      >
                        <Copy className='h-4 w-4' />
                      </Button>
                      <Button
                        variant='ghost'
                        size='icon'
                        title='Delete'
                        disabled={campaign.status === MailCampaignStatus.Sending}
                        onClick={() => setCampaignToDelete(campaign)}
                      >
                        <Trash2 className='h-4 w-4' />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <DeleteConfirmation
        isOpen={campaignToDelete !== null}
        setOpen={(open) => !open && setCampaignToDelete(null)}
        deleteMessage={`Are you sure you want to delete "${campaignToDelete?.name}"?`}
        onClick={handleDelete}
      />
    </div>
  )
}
