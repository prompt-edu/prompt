import DOMPurify from 'dompurify'

interface SanitizedHtmlProps {
  html: string
  className?: string
}

// Registered once, at module scope: DOMPurify hooks are global, so registering
// per render would stack duplicates on every re-render.
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node.tagName === 'A' && node.hasAttribute('href')) {
    node.setAttribute('target', '_blank')
    node.setAttribute('rel', 'noopener noreferrer')
  }
})

// The instructor-authored HTML renderer for this remote. A Module Federation
// remote cannot import core's FormDescriptionHTML, so this is a local
// equivalent; it belongs in @tumaet/prompt-ui-components long term.
export const SanitizedHtml = ({ html, className }: SanitizedHtmlProps) => {
  const sanitized = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    ALLOW_UNKNOWN_PROTOCOLS: false,
  })

  return (
    <>
      <style>
        {`
          .sanitizedHtml a { cursor: pointer; color: #1d4ed8; text-decoration: underline; }
          .sanitizedHtml ol { list-style: decimal; margin: 0 0 1rem 1.5rem; padding: 0; }
          .sanitizedHtml ul { list-style: disc; margin: 0 0 1rem 1.5rem; padding: 0; }
          .sanitizedHtml li { margin: 0.25rem 0; }
          .sanitizedHtml li p.text-node { margin: 0; }
          .sanitizedHtml p.text-node { margin: 0 0 1rem; }
        `}
      </style>
      <div
        className={className ? `sanitizedHtml ${className}` : 'sanitizedHtml'}
        dangerouslySetInnerHTML={{ __html: sanitized }}
      />
    </>
  )
}
