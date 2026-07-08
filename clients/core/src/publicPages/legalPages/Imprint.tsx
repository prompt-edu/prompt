import { useAuthStore } from '@tumaet/prompt-shared-state'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import DOMPurify from 'dompurify'
import { ArrowLeft } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

export default function ImprintPage() {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const [content, setContent] = useState('')

  DOMPurify.addHook('afterSanitizeAttributes', (node) => {
    // set all elements owning target to target=_blank
    if ('target' in node) {
      node.setAttribute('target', '_blank')
      node.setAttribute('rel', 'noopener')
    }
  })

  useEffect(() => {
    window.scrollTo(0, 0)
  }, [])

  useEffect(() => {
    fetch('/imprint.html')
      .then((res) => res.text())
      .then((res) => DOMPurify.sanitize(res))
      .then((res) => setContent(res))
  }, [])

  return (
    <div className='container mx-auto py-8 px-4'>
      <Card className='w-full max-w-4xl mx-auto'>
        <CardHeader className='relative'>
          <Button
            variant='ghost'
            size='icon'
            className='absolute left-4 top-4'
            onClick={() => navigate(user ? '/management' : '/')}
            aria-label='Go back'
          >
            <ArrowLeft className='h-4 w-4' />
          </Button>
          <CardTitle className='text-3xl font-bold text-center'>Imprint</CardTitle>
          <CardDescription className='text-center'>
            Legal information and disclaimers
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-6'>
          <style>
            {`
          .imprint a {
              cursor: pointer;
              color: #1d4ed8; /* Equivalent to text-blue-700 */
              text-decoration: underline; /* Always underlined */
          }
          .imprint a:hover {
              text-decoration: underline; /* Ensure underline on hover */
          }

          .imprint ol {
              list-style: decimal; /* Restore list-style for ordered lists */
              margin: 0 0 1rem 1.5rem; /* Adjust spacing around lists */
              padding: 0;
          }

          .imprint ol ol {
              list-style: lower-alpha; /* Nested ordered list with lower-alpha style */
              margin: 0.25rem 0 0.25rem 1rem; /* Adjust nested list margin */
          }

          .imprint ol ol ol {
              list-style: lower-roman; /* Nested nested ordered list with lower-roman style */
              margin: 0.25rem 0 0.25rem 1rem;
          }

          .imprint ul {
              list-style: disc; /* Restore list-style for unordered lists */
              margin: 0 0 1rem 1.5rem; /* Adjust spacing around lists */
              padding: 0;
          }

          .imprint ul ul {
              list-style: circle; /* Nested unordered list with circle style */
              margin: 0.25rem 0 0.25rem 1rem;
          }

          .imprint ul ul ul {
              list-style: square; /* Nested nested unordered list with square style */
              margin: 0.25rem 0 0.25rem 1rem;
          }

          .imprint li {
              margin: 0.25rem 0; /* Reduced spacing between list items */
          }

          .imprint li p.text-node {
              margin: 0; /* Remove extra margin from text nodes inside list items */
          }

          .imprint p.text-node {
              margin: 0 0 1rem; /* Add smaller spacing between paragraphs */
          }
        `}
          </style>
          <div className='imprint' dangerouslySetInnerHTML={{ __html: content }} />
        </CardContent>
      </Card>
    </div>
  )
}
