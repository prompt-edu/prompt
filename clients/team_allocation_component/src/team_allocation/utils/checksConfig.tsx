import { Languages, Laptop, Star } from 'lucide-react'
import type { StudentCheck } from '../interfaces/studentCheck'
import { UNKNOWN_SKILL_PROFICIENCY } from '../interfaces/tease/skillProficiency'

const isMissingProficiency = (value: string | undefined | null): boolean =>
  !value || value === UNKNOWN_SKILL_PROFICIENCY

// There are no checks for the tutor and student comments: the server does not send them yet
// (`ConvertCourseParticipationToTeaseStudent` always returns empty lists), so they would always fail.
export const checksConfig: StudentCheck[] = [
  // Devices
  {
    label: 'Devices',
    extractor: (s) => s.devices,
    isEmpty: (arr) => arr.length === 0,
    missingMessage: 'devices information',
    problemDescription: 'All students are missing information on which devices they have.',
    details: 'Ask about available devices. Export from the phase where collected.',
    category: 'devices',
    highLevelCategory: 'previous',
    icon: <Laptop className='h-4 w-4' />,
  },

  // ScoreLevel
  {
    label: 'Score Level',
    extractor: (s) => s.introCourseProficiency,
    isEmpty: isMissingProficiency,
    missingMessage: 'score level information',
    problemDescription: 'All students are missing score levels.',
    details: 'Assign score levels. Export from the phase where collected.',
    category: 'score',
    highLevelCategory: 'previous',
    icon: <Star className='h-4 w-4' />,
  },

  // SelfEvaluation
  {
    label: 'Self Evaluation',
    extractor: (s) => s.introSelfEvaluation,
    isEmpty: isMissingProficiency,
    missingMessage: 'self evaluation information',
    problemDescription: 'All students are missing self-assessment information.',
    details: 'Collect self-assessment from students. Export from the phase where collected.',
    category: 'score',
    highLevelCategory: 'previous',
    icon: <Star className='h-4 w-4' />,
  },

  // Language Proficiency
  {
    label: 'English Proficiency',
    extractor: (s) => s.languages?.find((l) => l.language === 'en')?.proficiency,
    isEmpty: (v) => !v,
    missingMessage: 'English proficiency levels',
    problemDescription: 'All students are missing English proficiency levels.',
    details:
      'Add application question with possible answers "A1/A2, B1/B2, C1/C2, Native" and export with access key "language_proficiency_english".',
    category: 'language',
    highLevelCategory: 'previous',
    icon: <Languages className='h-4 w-4' />,
  },
  {
    label: 'German Proficiency',
    extractor: (s) => s.languages?.find((l) => l.language === 'de')?.proficiency,
    isEmpty: (v) => !v,
    missingMessage: 'German proficiency levels',
    problemDescription: 'All students are missing German proficiency levels.',
    details:
      'Add application question with possible answers "A1/A2, B1/B2, C1/C2, Native" and export with access key "language_proficiency_german".',
    category: 'language',
    highLevelCategory: 'previous',
    icon: <Languages className='h-4 w-4' />,
  },
]
