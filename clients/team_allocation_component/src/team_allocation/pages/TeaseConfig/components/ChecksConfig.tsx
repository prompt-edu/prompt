import { Languages, Laptop, MessageSquare, Star } from 'lucide-react'
import type { StudentCheck } from '../../../interfaces/studentCheck'

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

  // Comments
  {
    label: 'Tutor Comments',
    extractor: (s) => s.tutorComments,
    isEmpty: (arr) => !arr || arr.length === 0,
    missingMessage: 'tutor comments',
    problemDescription: 'All students are missing comments from tutors.',
    details: 'Allow tutors to add feedback. Export from the phase where collected.',
    category: 'comments',
    highLevelCategory: 'previous',
    icon: <MessageSquare className='h-4 w-4' />,
  },
  {
    label: 'Student Comments',
    extractor: (s) => s.studentComments,
    isEmpty: (arr) => !arr || arr.length === 0,
    missingMessage: 'student comments',
    problemDescription: 'All students are missing comments from students.',
    details: 'Allow students to give feedback. Export from the phase where collected.',
    category: 'comments',
    highLevelCategory: 'previous',
    icon: <MessageSquare className='h-4 w-4' />,
  },

  // ScoreLevel
  {
    label: 'Score Level',
    extractor: (s) => s.introCourseProficiency,
    isEmpty: (v) => v === undefined || v === null,
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
    isEmpty: (v) => v === undefined || v === null,
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
