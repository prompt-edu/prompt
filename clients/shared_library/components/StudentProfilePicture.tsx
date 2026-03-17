import { Avatar, AvatarImage, AvatarFallback } from '@tumaet/prompt-ui-components'
import { getGravatarUrl } from '@/lib/getGravatarUrl'

type AvatarSize = 'lg' | 'md' | 'sm'

const sizeStyles: Record<AvatarSize, string> = {
  lg: 'h-24 w-24 text-xl',
  md: 'h-10 w-10 text-sm',
  sm: 'h-6 w-6 text-[0.7em]',
}

export function ProfilePicture({
  email,
  firstName,
  lastName,
  size = 'md',
  className = '',
}: {
  email: string
  firstName: string
  lastName: string
  size?: AvatarSize
  className?: string
}) {
  const initials = (firstName?.charAt(0) || 'N') + (lastName?.charAt(0) || 'A')

  return (
    <Avatar className={`${sizeStyles[size]} ${className}`}>
      <AvatarImage src={getGravatarUrl(email)} alt={lastName} />
      <AvatarFallback className='w-full h-full flex items-center justify-center font-bold'>
        {initials}
      </AvatarFallback>
    </Avatar>
  )
}
