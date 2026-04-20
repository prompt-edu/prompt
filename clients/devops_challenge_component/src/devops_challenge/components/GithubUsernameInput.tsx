import { useState, type ChangeEvent } from 'react'
import { useCreateRepository } from '../pages/hooks/useCreateRepository'
import {
  Button,
  Input,
  Alert,
  AlertDescription,
  AlertTitle,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Checkbox,
} from '@tumaet/prompt-ui-components'
import { Loader2, AlertCircle, User, ExternalLink } from 'lucide-react'

export const GithubUsernameInput = () => {
  const [error, setError] = useState<string | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)
  const repositoryMutation = useCreateRepository(setError)
  const [githubUsername, setGithubUsername] = useState('')
  const [hasGithubProfile, setHasGithubProfile] = useState(false)

  const validateGithubUsername = (username: string): boolean => {
    if (
      !/^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])|\.(?=[a-zA-Z0-9]))*$/.test(username.trim()) &&
      username.length > 0
    ) {
      setValidationError(
        'GitHub username can only contain letters, numbers, hyphens, and dots. Hypens and dots cannot be at the beginning or end of the username.',
      )
      return false
    }

    if (username.length > 39) {
      setValidationError('GitHub username cannot exceed 39 characters.')
      return false
    }

    setValidationError(null)
    return true
  }

  const handleUsernameChange = (e: ChangeEvent<HTMLInputElement>) => {
    const newUsername = e.target.value
    setGithubUsername(newUsername)
    validateGithubUsername(newUsername)
  }

  const isInputValid = githubUsername.length > 0 && !validationError && hasGithubProfile

  return (
    <Card className='w-full max-w-md mx-auto'>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <User className='w-6 h-6' />
          GitHub Repository Setup
        </CardTitle>
      </CardHeader>
      <CardContent className='space-y-6'>
        <Alert className='bg-black-50 border-black-200'>
          <AlertCircle className='h-4 w-4 text-black-600' />
          <AlertTitle className='text-black-800'>Getting Started</AlertTitle>
          <AlertDescription className='text-black-700'>
            For this challenge, you need a Github account. Enter your GitHub username to create a
            repository for this challenge. If you do not have one, please signup on:{' '}
            <a
              href={'https://github.com/signup'}
              target='_blank'
              rel='noopener noreferrer'
              className='text-primary hover:underline inline-flex items-center mr-1 truncate'
            >
              {'https://github.com/signup'}
              <ExternalLink className='h-3 w-3 flex-shrink-0 ml-1' />
            </a>
            <br />
            Make sure to enter your full name and use a picture of yourself in your Github account
            so that we can verify your identity.
          </AlertDescription>
        </Alert>

        <div className='space-y-4'>
          <div className='space-y-2'>
            <div className='relative'>
              <Input
                placeholder='GitHub username'
                value={githubUsername}
                onChange={handleUsernameChange}
                className={`pl-10 ${validationError ? 'border-red-500' : ''}`}
                aria-invalid={!!validationError}
                aria-describedby={validationError ? 'username-error' : undefined}
              />
              <User className='absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5' />
            </div>
            {validationError && (
              <p id='username-error' className='text-sm text-red-500'>
                {validationError}
              </p>
            )}
          </div>

          <div className='flex items-start space-x-2 mt-4'>
            <Checkbox
              id='github-profile-check'
              checked={hasGithubProfile}
              onCheckedChange={(checked) => setHasGithubProfile(checked as boolean)}
            />
            <label
              htmlFor='github-profile-check'
              className='text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer'
            >
              I confirm that my GitHub account has my full name and a profile picture set up.
            </label>
          </div>

          <Button
            onClick={() => repositoryMutation.mutate(githubUsername)}
            disabled={!isInputValid || repositoryMutation.isPending}
            className='w-full'
          >
            {repositoryMutation.isPending ? (
              <>
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                Creating Repository
              </>
            ) : (
              'Create Repository'
            )}
          </Button>
        </div>

        {repositoryMutation.isError && error && (
          <Alert variant='destructive'>
            <AlertCircle className='h-4 w-4' />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {repositoryMutation.isSuccess && (
          <Alert className='bg-green-50 border-green-200'>
            <AlertCircle className='h-4 w-4 text-green-600' />
            <AlertTitle className='text-green-800'>Repository Created</AlertTitle>
            <AlertDescription className='text-green-700'>
              Your repository has been created successfully. You can now start the challenge.
            </AlertDescription>
          </Alert>
        )}
      </CardContent>
    </Card>
  )
}
