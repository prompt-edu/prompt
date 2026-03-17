import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@tumaet/prompt-ui-components'
import { PropsWithChildren, ReactNode } from 'react'
import { ChevronRight } from 'lucide-react'

interface SettingsCardProps extends PropsWithChildren {
  icon: ReactNode
  title: string
  description: string
}

export function SettingsCard({ children, icon, title, description }: SettingsCardProps) {
  return (
    <Card className='w-full'>
      <Collapsible>
        <CollapsibleTrigger className='group w-full'>
          <CardHeader className='py-3 px-4'>
            <div className='flex items-center gap-3'>
              {/* Left side */}
              <div className='p-2 bg-gray-50 dark:bg-gray-900/30 rounded-lg text-gray-600 dark:text-gray-400'>
                {icon}
              </div>

              {/* Title / description */}
              <div className='flex-1 text-left'>
                <CardTitle className='text-lg text-gray-900 dark:text-gray-100'>{title}</CardTitle>
                <CardDescription className='mt-1 text-gray-600 dark:text-gray-400'>
                  {description}
                </CardDescription>
              </div>

              {/* Chevron */}
              <ChevronRight
                className='
                  h-4 w-4 text-gray-400
                  transition-transform duration-200
                  group-data-[state=open]:rotate-90
                '
              />
            </div>
          </CardHeader>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <hr />
          <div className='pt-4'>{children}</div>
        </CollapsibleContent>
      </Collapsible>
    </Card>
  )
}
