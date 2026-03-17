import { FormDescription } from '@tumaet/prompt-ui-components'
import DOMPurify from 'dompurify'
import { useFormContext } from 'react-hook-form'

interface FormDescriptionHTMLProps {
  htmlCode: string
}

export const FormDescriptionHTML = ({ htmlCode }: FormDescriptionHTMLProps) => {
  const formContext = useFormContext()

  // allowing _blank in links, but only with noopener
  DOMPurify.addHook('afterSanitizeAttributes', function (node) {
    // set all elements owning target to target=_blank
    if ('target' in node) {
      node.setAttribute('target', '_blank')
      node.setAttribute('rel', 'noopener')
    }
  })

  const sanitizedHtmlCode = DOMPurify.sanitize(htmlCode)
  const content = (
    <>
      <style>
        {`
          .descriptionText a {
              cursor: pointer;
              color: #1d4ed8; /* Equivalent to text-blue-700 */
              text-decoration: underline; /* Always underlined */
          }
          .descriptionText a:hover {
              text-decoration: underline; /* Ensure underline on hover */
          }

          .descriptionText ol {
              list-style: decimal; /* Restore list-style for ordered lists */
              margin: 0 0 1rem 1.5rem; /* Adjust spacing around lists */
              padding: 0;
          }

          .descriptionText ol ol {
              list-style: lower-alpha; /* Nested ordered list with lower-alpha style */
              margin: 0.25rem 0 0.25rem 1rem; /* Adjust nested list margin */
          }

          .descriptionText ol ol ol {
              list-style: lower-roman; /* Nested nested ordered list with lower-roman style */
              margin: 0.25rem 0 0.25rem 1rem;
          }

          .descriptionText ul {
              list-style: disc; /* Restore list-style for unordered lists */
              margin: 0 0 1rem 1.5rem; /* Adjust spacing around lists */
              padding: 0;
          }

          .descriptionText ul ul {
              list-style: circle; /* Nested unordered list with circle style */
              margin: 0.25rem 0 0.25rem 1rem;
          }

          .descriptionText ul ul ul {
              list-style: square; /* Nested nested unordered list with square style */
              margin: 0.25rem 0 0.25rem 1rem;
          }

          .descriptionText li {
              margin: 0.25rem 0; /* Reduced spacing between list items */
          }

          .descriptionText li p.text-node {
              margin: 0; /* Remove extra margin from text nodes inside list items */
          }

          .descriptionText p.text-node {
              margin: 0 0 1rem; /* Add smaller spacing between paragraphs */
          }
        `}
      </style>
      <div className='descriptionText' dangerouslySetInnerHTML={{ __html: sanitizedHtmlCode }} />
    </>
  )

  if (!formContext) {
    return <div className='text-sm text-muted-foreground'>{content}</div>
  }

  return <FormDescription>{content}</FormDescription>
}
