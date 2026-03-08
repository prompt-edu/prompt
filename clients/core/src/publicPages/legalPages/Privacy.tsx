import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { ArrowLeft } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import DOMPurify from 'dompurify'
import { useAuthStore } from '@tumaet/prompt-shared-state'

export default function PrivacyPage() {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const [content, setContent] = useState('')

  DOMPurify.addHook('afterSanitizeAttributes', function (node) {
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
    fetch('/privacy.html')
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
          <CardTitle className='text-3xl font-bold text-center'>
            Datenschutzerklärung / Privacy Policy
          </CardTitle>
          <CardDescription className='text-center'>
            Informationen zum Datenschutz und Ihren Rechten / Information about data protection and
            your rights
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-6'>
          <style>
            {`
          .privacy a {
              cursor: pointer;
              color: #1d4ed8; /* Equivalent to text-blue-700 */
              text-decoration: underline; /* Always underlined */
          }
          .privacy a:hover {
              text-decoration: underline; /* Ensure underline on hover */
          }

          .privacy ol {
              list-style: decimal; /* Restore list-style for ordered lists */
              margin: 0 0 1rem 1.5rem; /* Adjust spacing around lists */
              padding: 0;
          }

          .privacy ol ol {
              list-style: lower-alpha; /* Nested ordered list with lower-alpha style */
              margin: 0.25rem 0 0.25rem 1rem; /* Adjust nested list margin */
          }

          .privacy ol ol ol {
              list-style: lower-roman; /* Nested nested ordered list with lower-roman style */
              margin: 0.25rem 0 0.25rem 1rem;
          }

          .privacy ul {
              list-style: disc; /* Restore list-style for unordered lists */
              margin: 0 0 1rem 1.5rem; /* Adjust spacing around lists */
              padding: 0;
          }

          .privacy ul ul {
              list-style: circle; /* Nested unordered list with circle style */
              margin: 0.25rem 0 0.25rem 1rem;
          }

          .privacy ul ul ul {
              list-style: square; /* Nested nested unordered list with square style */
              margin: 0.25rem 0 0.25rem 1rem;
          }

          .privacy li {
              margin: 0.25rem 0; /* Reduced spacing between list items */
          }

          .privacy li p.text-node {
              margin: 0; /* Remove extra margin from text nodes inside list items */
          }

          .privacy p.text-node {
              margin: 0 0 1rem; /* Add smaller spacing between paragraphs */
          }
        `}
          </style>
          <div className='privacy' dangerouslySetInnerHTML={{ __html: content }} />
        </CardContent>
      </Card>
    </div>
  )
}
