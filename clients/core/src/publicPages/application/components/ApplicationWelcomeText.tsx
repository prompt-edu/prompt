import { FormDescriptionHTML } from '../pages/ApplicationForm/components/FormDescriptionHTML'

interface ApplicationWelcomeTextProps {
  welcomeText?: string | null
}

export const ApplicationWelcomeText = ({ welcomeText }: ApplicationWelcomeTextProps) => {
  if (!welcomeText) {
    return null
  }

  return (
    <div className='max-w-3xl mx-auto text-left' data-testid='application-welcome-text-content'>
      <FormDescriptionHTML
        htmlCode={welcomeText}
        className='text-base text-muted-foreground leading-relaxed'
      />
    </div>
  )
}
