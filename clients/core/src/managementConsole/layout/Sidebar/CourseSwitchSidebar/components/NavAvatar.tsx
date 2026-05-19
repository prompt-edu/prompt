import { Avatar, AvatarImage, AvatarFallback } from '@tumaet/prompt-ui-components'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import { getGravatarUrl } from '@tumaet/prompt-shared-state'

interface NavAvatarProps {
  avatarOnly?: boolean
}

export function NavAvatar({ avatarOnly = false }: NavAvatarProps) {
  const { user } = useAuthStore()

  const userName = user?.firstName + ' ' + user?.lastName || ' Unknown User'
  const userEmail = user?.email || 'Unknown Email'
  const initials = `${user?.firstName.charAt(0)}${user?.lastName.charAt(0)}` || '??'

  return (
    <>
      <Avatar className='h-10 w-10 rounded-lg'>
        <AvatarImage src={getGravatarUrl(userEmail)} alt={userName} />
        <AvatarFallback className='rounded-lg'>{initials}</AvatarFallback>
      </Avatar>
      {!avatarOnly && (
        <div className='grid flex-1 text-left text-sm leading-tight'>
          <span className='truncate font-semibold'>{userName}</span>
          <span className='truncate text-xs'>{userEmail}</span>
        </div>
      )}
    </>
  )
}
