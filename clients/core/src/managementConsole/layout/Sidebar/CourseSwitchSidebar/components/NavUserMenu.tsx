import { LogOut, Shield } from 'lucide-react'

import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@tumaet/prompt-ui-components'
import { NavAvatar } from './NavAvatar'
import { ThemeToggle } from '@tumaet/prompt-ui-components'
import { useNavigate } from 'react-router-dom'

interface NavUserProps {
  onLogout: () => void
  showThemeToggle?: boolean
}

export function NavUserMenu({ onLogout, showThemeToggle = true }: NavUserProps) {
  const navigate = useNavigate()
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className='max-w-10 max-h-10 min-h-10 min-w-10 p-0'>
          <NavAvatar avatarOnly />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        className='w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg'
        side='bottom'
        align='end'
        sideOffset={4}
      >
        <DropdownMenuLabel className='p-0 font-normal'>
          <div className='flex items-center gap-2 px-1 py-1.5 text-left text-sm'>
            <NavAvatar />
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {showThemeToggle && (
          <div className='p-2'>
            <p className='text-sm text-muted-foreground mb-3'>Preferences</p>
            <div className='flex flex-col gap-4'>
              <div className='flex items-center justify-between'>
                <span className='text-sm'>Theme</span>
                <ThemeToggle />
              </div>
            </div>
          </div>
        )}
        {showThemeToggle && <DropdownMenuSeparator />}
        <DropdownMenuItem onClick={() => navigate('/management/privacy')}>
          <Shield />
          Privacy
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onLogout}>
          <LogOut />
          Log out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
