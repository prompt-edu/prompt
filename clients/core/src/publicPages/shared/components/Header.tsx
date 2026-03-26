import { Button } from '@tumaet/prompt-ui-components'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import { LogIn } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import packageJSON from '../../../../package.json'
import { NavUserMenu } from '@managementConsole/layout/Sidebar/CourseSwitchSidebar/components/NavUserMenu'

interface HeaderProps {
  withLoginButton?: boolean
  onLogout?: () => void
}

export const Header = ({ withLoginButton = true, onLogout }: HeaderProps) => {
  const navigate = useNavigate()
  const version = packageJSON.version
  const { user } = useAuthStore()

  return (
    <div className='flex flex-col sm:flex-row justify-between items-center mb-12 gap-4'>
      <div className='flex items-center space-x-4'>
        <img src='/prompt_logo.svg' alt='PROMPT Logo' className='h-12 w-12' />
        <div className='relative flex items-baseline'>
          <span className='text-2xl font-extrabold tracking-wide text-primary drop-shadow-sm'>
            PROMPT
          </span>
          <span className='ml-1 text-s font-normal text-gray-400'>{version}</span>
        </div>
      </div>
      {withLoginButton && !user && (
        <Button
          variant='outline'
          className='flex items-center space-x-2'
          onClick={() => navigate('/management')}
        >
          <LogIn className='h-4 w-4' />
          <span>Login</span>
        </Button>
      )}
      {user && (
        <div className='flex items-center gap-2'>
          <Button
            variant='outline'
            className='flex items-center space-x-2'
            onClick={() => navigate('/management')}
          >
            <LogIn className='h-4 w-4' />
            <span>Go to App</span>
          </Button>
          <NavUserMenu onLogout={onLogout ?? (() => {})} showThemeToggle={false} />
        </div>
      )}
    </div>
  )
}
