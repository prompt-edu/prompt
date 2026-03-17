import { ArrowUpRight } from 'lucide-react'
import { PropsWithChildren } from 'react'
import { Link } from 'react-router-dom'

interface LinkHeadingProps extends PropsWithChildren {
  targetURL: string
}

export const LinkHeading = ({ children, targetURL }: LinkHeadingProps) => {
  return (
    <div className='hover:text-blue-500'>
      <Link to={targetURL} className='flex items-center'>
        {children}
        <ArrowUpRight className='ml-1 w-4 h-4 ' />
      </Link>
    </div>
  )
}
