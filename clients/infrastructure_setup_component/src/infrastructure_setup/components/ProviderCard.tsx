import { useMutation } from '@tanstack/react-query'
import { Badge, Button, Card, CardContent, useToast } from '@tumaet/prompt-ui-components'
import { CheckCircle2, Pencil, ShieldCheck } from 'lucide-react'
import axios from 'axios'

import { ProviderConfig } from '../interfaces/providerConfig'
import { validateProviderConfig } from '../network/mutations/validateProviderConfig'

interface Props {
  coursePhaseID: string
  provider: ProviderConfig
  onEdit: (provider: ProviderConfig) => void
}

export const ProviderCard = ({ coursePhaseID, provider, onEdit }: Props) => {
  const { toast } = useToast()

  const { mutate: validate, isPending: isValidating } = useMutation({
    mutationFn: () => validateProviderConfig(coursePhaseID, provider.providerType),
    onSuccess: () => {
      toast({
        title: 'Credentials valid',
        description: `${provider.providerType} credentials accepted by the provider.`,
      })
    },
    onError: (err: unknown) => {
      const description =
        axios.isAxiosError(err) && err.response?.data?.error
          ? String(err.response.data.error)
          : err instanceof Error
            ? err.message
            : 'Validation failed.'
      toast({
        title: 'Validation failed',
        description,
        variant: 'destructive',
      })
    },
  })

  return (
    <Card>
      <CardContent className='flex items-center justify-between gap-4 p-4'>
        <div className='flex items-center gap-3'>
          <ShieldCheck className='h-5 w-5 text-blue-500' />
          <div>
            <div className='flex items-center gap-2'>
              <p className='font-medium capitalize'>{provider.providerType}</p>
              <Badge variant='secondary' className='gap-1'>
                <CheckCircle2 className='h-3 w-3' /> configured
              </Badge>
            </div>
            <p className='font-mono text-xs text-muted-foreground'>{provider.id}</p>
          </div>
        </div>

        <div className='flex items-center gap-2'>
          <Button variant='outline' size='sm' onClick={() => validate()} disabled={isValidating}>
            {isValidating ? 'Validating…' : 'Validate'}
          </Button>
          <Button variant='outline' size='sm' onClick={() => onEdit(provider)}>
            <Pencil className='mr-1 h-3 w-3' /> Edit
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

export default ProviderCard
