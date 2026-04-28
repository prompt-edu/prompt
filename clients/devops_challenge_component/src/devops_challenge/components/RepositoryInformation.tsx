import { useState } from 'react'
import { useGetDeveloperProfile } from '../pages/hooks/useGetDeveloperProfile'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Button,
  Alert,
  AlertDescription,
  AlertTitle,
} from '@tumaet/prompt-ui-components'
import { FolderGit2, ExternalLink, Clipboard, Check, AlertCircle } from 'lucide-react'

export const RepositoryInformation = () => {
  const developerQuery = useGetDeveloperProfile()
  const repoUrl = developerQuery.data?.repositoryUrl || ''

  const [copied, setCopied] = useState(false)

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(repoUrl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000) // Reset after 2 seconds
    } catch (err) {
      console.error('Failed to copy: ', err)
    }
  }

  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='text-xl flex items-center'>
          <FolderGit2 className='mr-2 h-5 w-5' />
          Repository Information
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className='space-y-4'>
          <div className='flex flex-col space-y-1'>
            <div className='space-y-4'>
              <p className='text-slate-700 leading-relaxed'>
                We have created a repository for you. Only you have access to this repository.
              </p>

              <Alert variant='default'>
                <AlertCircle className='h-4 w-4 text-amber-600' />
                <AlertTitle className='text-amber-800 font-medium'>Important</AlertTitle>
                <AlertDescription className='text-amber-700'>
                  You must solve this challenge independently! No collaboration with other students
                  is allowed.
                </AlertDescription>
              </Alert>

              <div className='space-y-1'>
                <p className='text-slate-700 leading-relaxed'>
                  You may use GenAI tools such as Github Co-Pilot or ChatGPT, but make sure that you
                  understand what you are doing!
                </p>

                <ol className='list-decimal pl-5 text-slate-700 space-y-2 my-4'>
                  <li>Open the GitHub repository using the link below</li>
                  <li>Carefully read the Challenge Description in the README</li>
                  <li>Clone the repository</li>
                  <li>Solve the challenge</li>
                  <li>Commit and push your changes</li>
                </ol>

                <p className='text-slate-700 leading-relaxed'>
                  When you think you have successfully solved the challenge, click on Start Testing.
                  <br />
                  You have at most 3 attempts to test your solution, so use them wisely.
                </p>
              </div>
            </div>
            <div className='h-2 bg-slate-200' />
            <h3 className='text-lg font-semibold mb-2 flex items-center'>
              <FolderGit2 className='mr-2 h-5 w-5 text-slate-700' />
              Repository URL
            </h3>
            <div className='flex items-center gap-2'>
              <div className='flex-1 flex items-center'>
                <a
                  href={repoUrl}
                  target='_blank'
                  rel='noopener noreferrer'
                  className='text-primary hover:underline mr-1 truncate'
                >
                  {repoUrl}
                </a>
                <ExternalLink className='h-3 w-3 shrink-0' />
              </div>
              <Button
                variant='ghost'
                size='icon'
                className='h-8 w-8 rounded-full'
                onClick={copyToClipboard}
                title='Copy Repository URL'
                aria-label='Copy Repository URL'
              >
                {copied ? (
                  <Check className='h-4 w-4 text-green-500' />
                ) : (
                  <Clipboard className='h-4 w-4' />
                )}
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
