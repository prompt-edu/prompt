import { env } from '@tumaet/prompt-shared-state'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import packageJSON from '../../../../../../package.json'

interface GitHubPR {
  title: string
  labels: { name: string }[]
}

export const PromptLogo = () => {
  const version = packageJSON.version
  const prMatch = env.GITHUB_REF.match(/refs\/pull\/(\d+)\/merge/)
  const pr = prMatch?.[1]

  const { data } = useQuery<GitHubPR>({
    queryKey: ['github-pr', pr],
    queryFn: () =>
      fetch(`https://api.github.com/repos/prompt-edu/prompt/pulls/${pr}`).then((r) => r.json()),
    staleTime: Infinity,
    enabled: !!pr,
  })

  const hasMigration = data?.labels?.some((l) => l.name.toLowerCase().includes('migration'))

  const content = (
    <>
      <img
        src='/prompt_logo.svg'
        alt='Prompt logo'
        className={`size-8 -mr-1 ${hasMigration ? 'filter-[sepia(1)_saturate(5)_hue-rotate(330deg)]' : ''}`}
      />
      <div className='relative flex items-baseline'>
        <span className='text-lg font-extrabold tracking-wide text-primary drop-shadow-xs'>
          PROMPT
        </span>
        <span className='ml-1 text-xs font-normal text-gray-400'>{pr ? `#${pr}` : version}</span>
      </div>
    </>
  )

  if (pr) {
    return (
      <a
        href={`https://github.com/prompt-edu/prompt/pull/${pr}`}
        target='_blank'
        rel='noopener noreferrer'
        className='flex items-center hover:opacity-80 transition-opacity duration-150'
      >
        {content}
      </a>
    )
  }

  return (
    <Link to='/' className='flex items-center hover:opacity-80 transition-opacity duration-150'>
      {content}
    </Link>
  )
}
