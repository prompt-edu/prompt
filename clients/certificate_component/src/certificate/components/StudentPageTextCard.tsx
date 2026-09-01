import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  DescriptionMinimalTiptapEditor,
  TooltipProvider,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { updateStudentPageText } from '../network/queries/getConfig'

interface StudentPageTextCardProps {
  phaseId: string
  initialText: string
}

// An emptied rich text editor still emits markup like `<p></p>`, which must count
// as no text at all. The value is parsed rather than stripped with a regex: this
// only decides emptiness (sanitizing happens where the HTML is rendered), and a
// regex tag-strip reads like - and gets flagged as - a broken sanitizer.
const RENDERS_WITHOUT_TEXT = 'img, hr, iframe, video, audio, table'

const isEmptyRichText = (html: string) => {
  if (typeof DOMParser === 'undefined') {
    return html.trim() === ''
  }
  // A parsed document is inert: nothing in it loads or runs.
  const { body } = new DOMParser().parseFromString(html, 'text/html')
  return body.textContent?.trim() === '' && body.querySelector(RENDERS_WITHOUT_TEXT) === null
}

const normalizeHtml = (html: string) => (isEmptyRichText(html) ? '' : html)

const errorMessage = (error: unknown): string | undefined => {
  const response = (error as { response?: { data?: { error?: unknown } } })?.response
  return typeof response?.data?.error === 'string' ? response.data.error : undefined
}

// Mounted only once the config has loaded, so the editor - which reads its
// content once, when it is created - starts from the stored text.
export const StudentPageTextCard = ({ phaseId, initialText }: StudentPageTextCardProps) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()

  const [text, setText] = useState(initialText)
  const [savedText, setSavedText] = useState(initialText)

  // The editor re-serializes on create (adding its own classes), so its first
  // emission is the baseline to compare against, not an edit. Any emission after
  // mount has settled is a real change.
  const hydrating = useRef(true)
  useEffect(() => {
    hydrating.current = false
  }, [])

  const mutation = useMutation({
    mutationFn: (value: string | null) => updateStudentPageText(phaseId, value),
    onSuccess: (_, value) => {
      setSavedText(value ?? '')
      queryClient.invalidateQueries({ queryKey: ['config', phaseId] })
      toast({ title: 'Success', description: 'Student page text updated successfully' })
    },
    onError: (error) => {
      toast({
        title: 'Error',
        description: errorMessage(error) ?? 'Failed to update student page text',
        variant: 'destructive',
      })
    },
  })

  const handleChange = (value: unknown) => {
    const html = typeof value === 'string' ? value : ''
    setText(html)
    if (hydrating.current) {
      hydrating.current = false
      setSavedText(normalizeHtml(html))
    }
  }

  const normalized = normalizeHtml(text)
  const hasChanges = normalized !== savedText

  return (
    <Card data-testid='certificate-student-page-text-settings'>
      <CardHeader>
        <CardTitle>Student Download Page Text</CardTitle>
        <CardDescription>
          An optional message shown to students on their certificate page, for example how to add
          the certificate to a profile or who to contact with questions. Leave it empty to show
          nothing extra.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <TooltipProvider>
          <DescriptionMinimalTiptapEditor
            value={text}
            onChange={handleChange}
            className='w-full'
            editorContentClassName='p-3'
            output='html'
            placeholder='Type the message students should see...'
            autofocus={false}
            editable={true}
            editorClassName='focus:outline-hidden'
          />
        </TooltipProvider>
        <div className='flex justify-end'>
          <Button
            data-testid='certificate-student-page-text-save'
            onClick={() => mutation.mutate(normalized === '' ? null : normalized)}
            disabled={mutation.isPending || !hasChanges}
          >
            {mutation.isPending && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            Save Student Page Text
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
