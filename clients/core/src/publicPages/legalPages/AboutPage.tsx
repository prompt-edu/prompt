import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Separator,
} from '@tumaet/prompt-ui-components'
import {
  ArrowLeft,
  Bug,
  GitPullRequest,
  Mail,
  FileText,
  FileUser,
  UserCheck,
  Users,
  Mic,
  Upload,
  Plus,
  Layers,
} from 'lucide-react'
import { Link, useNavigate } from 'react-router-dom'
import { ContributorList } from './components/ContributorList'
import { env } from '@tumaet/prompt-shared-state'
import { useAuthStore } from '@tumaet/prompt-shared-state'

export default function AboutPage() {
  const navigate = useNavigate()
  const { user } = useAuthStore()

  const coreFeatures = [
    {
      icon: Layers,
      title: 'Course Configuration',
      description:
        'Build a course by assembling various course phases to suit your specific teaching needs.',
    },
    {
      icon: FileUser,
      title: 'Application Phase',
      description:
        'Streamline the application process for courses, making it easier for students to apply and for instructors to manage applications.',
    },
    {
      icon: UserCheck,
      title: 'Student Management',
      description: 'Efficiently manage student information and course participation.',
    },
  ]

  const dynamicPhases = [
    {
      icon: Mic,
      title: 'Interview Phase',
      description:
        'Manage and schedule interviews with applicants as part of the selection process.',
    },
    {
      icon: Upload,
      title: 'TUM Matching Export',
      description: 'Export data in a format compatible with TUM Matching for seamless integration.',
    },
    {
      icon: Users,
      title: 'Team Phase',
      description:
        'Assign students to teams and projects, and manage project work throughout the course.',
    },
    {
      icon: UserCheck,
      title: 'Assessment Phase',
      description:
        'Conduct structured peer, self, and instructor assessments using a configurable assessment framework.',
    },
  ]

  return (
    <div className='container mx-auto py-12 px-4'>
      <Card className='w-full max-w-6xl mx-auto shadow-lg'>
        <CardHeader className='relative pb-8'>
          <Button
            variant='ghost'
            size='icon'
            className='absolute left-4 top-4 hover:bg-gray-100 transition-colors'
            onClick={() => navigate(user ? '/management' : '/')}
            aria-label='Go back'
          >
            <ArrowLeft className='h-5 w-5' />
          </Button>
          <CardTitle className='text-4xl font-bold text-center mt-8'>About PROMPT</CardTitle>
          <CardDescription className='text-center text-lg mt-2'>
            Learn more about our robust course management platform
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-12'>
          <section>
            <h2 className='text-2xl font-semibold mb-4'>What is PROMPT?</h2>
            <p className='text-gray-700 leading-relaxed'>
              PROMPT (Project-Oriented Modular Platform for Teaching) is a course management tool
              specifically designed for project-based university courses. By supporting a wide range
              of organizational processes, it reduces the administrative burden typically associated
              with such courses and aims to streamline the daily activities of both students and
              instructors, enhancing the overall learning experience.
            </p>
            <p className='text-gray-700 leading-relaxed'>
              Originally developed for the iPraktikum at the Technical University of Munich, PROMPT
              has been reimagined with a flexible, modular architecture. Each course is built from
              independent, reusable components that can be easily extended, giving instructors the
              freedom to tailor functionality and structure to their exact teaching needs.
            </p>
          </section>

          <section>
            <h2 className='text-2xl font-semibold mb-4'>Get in Touch</h2>
            <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
              {[
                {
                  icon: Bug,
                  text: 'Report a bug',
                  link: 'https://github.com/prompt-edu/prompt/issues',
                },
                {
                  icon: GitPullRequest,
                  text: 'Request a feature',
                  link: 'https://github.com/prompt-edu/prompt/issues',
                },
                { icon: Mail, text: 'Contact us', link: 'https://ase.cit.tum.de/impressum/' },
                {
                  icon: FileText,
                  text: 'Release notes',
                  link: 'https://github.com/prompt-edu/prompt/releases',
                },
              ].map((item, index) => (
                <a
                  key={index}
                  href={item.link}
                  className='flex items-center p-3 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors'
                  target='_blank'
                  rel='noopener noreferrer'
                >
                  <item.icon className='h-5 w-5 mr-3 text-blue-600' />
                  <span className='text-gray-700 hover:text-gray-900'>{item.text}</span>
                </a>
              ))}
            </div>
          </section>

          <Separator />

          <section>
            <h2 className='text-2xl font-semibold mb-3'>Main Features</h2>
            <p className='text-gray-700 leading-relaxed'>
              The core features are built-in functionalities essential for course management, while
              dynamically loaded phases are additional, customizable components that can be added as
              needed.
            </p>

            <h3 className='text-xl font-semibold mt-8 mb-3'>Core Features</h3>
            <h4 className='text-l mb-4 text-secondary-foreground'>
              The Core offers a range of essential features designed to enhance the efficiency and
              effectiveness of course management.
            </h4>
            <div className='grid grid-cols-1 md:grid-cols-3 gap-6 mb-8'>
              {coreFeatures.map((feature, index) => (
                <Card key={index}>
                  <CardHeader>
                    <CardTitle className='text-lg font-semibold flex items-center'>
                      <feature.icon className='h-5 w-5 mr-2 text-blue-600' />
                      {feature.title}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className='text-gray-700 text-sm'>{feature.description}</p>
                  </CardContent>
                </Card>
              ))}
            </div>

            <h3 className='text-xl font-semibold mb-3'>Dynamically Loaded Course Phases</h3>
            <h4 className='text-l mb-4 text-secondary-foreground'>
              PROMPT allows instructors to create and manage own independent course phases,
              fostering a collaborative and easily extensible platform for project-based learning.
            </h4>

            <div className='grid grid-cols-1 md:grid-cols-2 md:grid-rows-2 gap-6'>
              {dynamicPhases.map((phase, index) => (
                <Card key={index}>
                  <CardHeader>
                    <CardTitle className='text-lg font-semibold flex items-center'>
                      <phase.icon className='h-5 w-5 mr-2 text-blue-600' />
                      {phase.title}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className='text-gray-700 text-sm'>{phase.description}</p>
                  </CardContent>
                </Card>
              ))}
              <Card className='border border-dashed border-blue-600'>
                <CardHeader>
                  <CardTitle className='text-lg font-semiold flex items-center'>
                    <Plus className='h-5 w-5 mr-2 text-blue-600' />
                    Custom Course Phase
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className='text-gray-700 text-sm'>
                    Easily extend PROMPT with custom phases tailored to your course needs.
                  </p>
                  <div className='flex justify-end mt-3'>
                    <Button variant='outline'>
                      <Link to='https://prompt-edu.github.io/prompt/' className='btn btn-outline'>
                        Learn More
                      </Link>
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          </section>

          <Separator />

          <section>
            <h2 className='text-2xl font-semibold mb-6'>Contributors</h2>
            <ContributorList />
          </section>

          <Separator />

          <section>
            <h2 className='text-2xl font-semibold mb-6'>Release Info</h2>
            <ul className='list-disc list-inside text-gray-700'>
              <li>
                <span className='font-semibold'>Github SHA:</span> {env.GITHUB_SHA}
              </li>
              <li>
                <span className='font-semibold'>Github Ref:</span> {env.GITHUB_REF}
              </li>
              <li>
                <span className='font-semibold'>Image Tag:</span> {env.SERVER_IMAGE_TAG}
              </li>
              <li>
                <span className='font-semibold'>License:</span> MIT
              </li>
            </ul>
          </section>
        </CardContent>
      </Card>
    </div>
  )
}
