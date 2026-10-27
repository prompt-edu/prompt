import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Separator,
  Skeleton,
} from '@tumaet/prompt-ui-components'
import { CheckCircle2, Server, XCircle } from 'lucide-react'
import type { ReactNode } from 'react'
import type { ClientInfo } from '../interfaces/clientInfo'
import { CAPABILITY_LABELS, type ServiceInfo } from '../interfaces/serviceCapabilities'
import {
  clientStatus,
  type PhaseStatusEntry,
  type ProbeResult,
  phaseCardStatus,
  type ServiceStatus,
  serverStatus,
} from '../utils/phaseStatus'
import { ServiceStatusBadge } from './ServiceStatusBadge'

interface ServiceStatusCardProps {
  entry: PhaseStatusEntry
  server?: ProbeResult<ServiceInfo>
  client?: ProbeResult<ClientInfo>
}

interface StatusSectionProps {
  title: string
  probe: ProbeResult<unknown>
  version: string | undefined
  status: ServiceStatus
}

const ALL_KNOWN_CAPABILITIES = Object.keys(CAPABILITY_LABELS)

const HEALTH: Record<ServiceStatus, { label: string; className: string }> = {
  Online: { label: 'Healthy', className: 'text-green-600' },
  OnlineUnhealthy: { label: 'Degraded', className: 'text-yellow-600' },
  Offline: { label: 'Unreachable', className: 'text-red-600' },
}

const SectionTitle = ({ children }: { children: ReactNode }) => (
  <p className='text-xs font-semibold uppercase tracking-wide text-muted-foreground'>{children}</p>
)

const StatusSection = ({ title, probe, version, status }: StatusSectionProps) => (
  <div className='flex flex-col gap-1.5'>
    <SectionTitle>{title}</SectionTitle>
    {probe.isPending ? (
      <div className='flex flex-col gap-2'>
        <Skeleton className='h-4 w-1/2' />
        <Skeleton className='h-4 w-1/3' />
      </div>
    ) : (
      <dl className='grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm'>
        <dt className='text-muted-foreground'>Version</dt>
        <dd>{version ?? 'n/a'}</dd>
        <dt className='text-muted-foreground'>Health</dt>
        <dd className={HEALTH[status].className}>{HEALTH[status].label}</dd>
      </dl>
    )}
  </div>
)

export const ServiceStatusCard = ({ entry, server, client }: ServiceStatusCardProps) => {
  const serverRowStatus = server && serverStatus(server)
  const clientRowStatus = client && clientStatus(client)
  const cardStatus = phaseCardStatus(server, client)
  const capabilities = server && !server.isError ? server.data?.capabilities : undefined

  return (
    <Card>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between gap-2'>
          <div className='flex min-w-0 items-center gap-2'>
            <Server className='h-4 w-4 shrink-0 text-muted-foreground' />
            <CardTitle
              className='line-clamp-2 min-w-0 wrap-anywhere text-base'
              title={entry.phaseType.name}
            >
              {entry.phaseType.name}
            </CardTitle>
          </div>
          {cardStatus ? (
            <ServiceStatusBadge status={cardStatus} />
          ) : (
            <Skeleton className='h-5 w-16' />
          )}
        </div>
      </CardHeader>
      <CardContent className='flex flex-col gap-4'>
        {server && serverRowStatus && (
          <StatusSection
            title='Server'
            probe={server}
            version={server.isError ? undefined : server.data?.version || undefined}
            status={serverRowStatus}
          />
        )}
        {client && clientRowStatus && (
          <StatusSection
            title='Client'
            probe={client}
            version={client.isError ? undefined : client.data?.buildVersion}
            status={clientRowStatus}
          />
        )}
        {capabilities && (
          <>
            <Separator />
            <Capabilities capabilities={capabilities} />
          </>
        )}
      </CardContent>
    </Card>
  )
}

const Capabilities = ({ capabilities }: { capabilities: Record<string, boolean> }) => (
  <div className='flex flex-col gap-1.5'>
    <SectionTitle>Capabilities</SectionTitle>
    <ul className='flex flex-col gap-1.5'>
      {ALL_KNOWN_CAPABILITIES.map((key) => {
        const supported = capabilities[key] === true
        return (
          <li key={key} className='flex items-center gap-2 text-sm'>
            {supported ? (
              <CheckCircle2 className='h-4 w-4 shrink-0 text-green-500' />
            ) : (
              <XCircle className='h-4 w-4 shrink-0 text-muted-foreground' />
            )}
            <span className={supported ? '' : 'text-muted-foreground'}>
              {CAPABILITY_LABELS[key]}
            </span>
          </li>
        )
      })}
    </ul>
  </div>
)
