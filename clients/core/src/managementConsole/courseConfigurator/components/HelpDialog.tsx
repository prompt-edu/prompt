import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@tumaet/prompt-ui-components'
import { HelpCircle, Lightbulb, MousePointer, ArrowRight } from 'lucide-react'
import { useState } from 'react'

export const HelpDialog = () => {
  const [helpOpen, setHelpOpen] = useState(false)

  return (
    <Dialog open={helpOpen} onOpenChange={setHelpOpen}>
      <DialogTrigger asChild>
        <button
          className={
            'flex items-center gap-2 px-3 py-2 rounded-lg bg-gray-50 hover:bg-gray-100 text-gray-700 ' +
            'dark:bg-gray-800 dark:hover:bg-gray-700 dark:text-gray-200 font-medium border border-gray-200 ' +
            'dark:border-gray-600 transition-colors'
          }
          aria-label='Open Help Dialog'
        >
          <HelpCircle className='h-4 w-4' />
          Help
        </button>
      </DialogTrigger>
      <DialogContent className='max-w-2xl max-h-[80vh] overflow-y-auto'>
        <DialogHeader>
          <DialogTitle className='text-xl font-semibold flex items-center gap-2 text-gray-900 dark:text-gray-100'>
            <HelpCircle className='h-5 w-5 text-gray-600 dark:text-gray-300' />
            Course Configurator Help
          </DialogTitle>
          <p className='text-gray-600 dark:text-gray-400 text-sm mt-1'>
            Learn how to set up your course using the visual editor.
          </p>
        </DialogHeader>

        <div className='space-y-6 mt-6'>
          {/* Overview */}
          <div>
            <h3 className='text-lg font-medium mb-4 flex items-center gap-2 text-gray-900 dark:text-gray-100'>
              <Lightbulb className='h-4 w-4 text-gray-600 dark:text-gray-300' />
              What is the Course Configurator?
            </h3>
            <p className='text-gray-700 dark:text-gray-300 text-sm leading-relaxed'>
              The configurator lets you build and customize your course using drag-and-drop. A
              course consists of multiple phases that you can arrange and connect to control both
              student and data flow.
            </p>
          </div>

          {/* How to Use */}
          <div>
            <h3 className='text-lg font-medium mb-4 flex items-center gap-2 text-gray-900 dark:text-gray-100'>
              <MousePointer className='h-4 w-4 text-gray-600 dark:text-gray-300' />
              Getting Started
            </h3>

            <div className='space-y-3'>
              {/* Step 1 */}
              <div className='flex gap-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700'>
                <div className='flex items-center justify-center w-6 h-6 bg-gray-100 dark:bg-gray-700 rounded-full shrink-0 mt-0.5'>
                  <span className='text-xs font-medium text-gray-600 dark:text-gray-300'>1</span>
                </div>
                <div>
                  <h4 className='font-medium text-gray-900 dark:text-gray-100 mb-1'>Add Phases</h4>
                  <p className='text-gray-600 dark:text-gray-300 text-sm leading-relaxed'>
                    Use the sidebar to browse available course phases. Drag phases onto the canvas
                    to add them to your course.
                  </p>
                </div>
              </div>

              {/* Step 2 */}
              <div className='flex gap-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700'>
                <div className='flex items-center justify-center w-6 h-6 bg-blue-100 dark:bg-blue-900 rounded-full shrink-0 mt-0.5'>
                  <div className='text-xs font-medium text-blue-600 dark:text-blue-300'>2</div>
                </div>
                <div>
                  <h4 className='font-medium text-gray-900 dark:text-gray-100 mb-1'>
                    Connect Student Flow
                  </h4>
                  <p className='text-gray-600 dark:text-gray-300 text-sm leading-relaxed'>
                    Define the order students go through phases by connecting the{' '}
                    <span className='inline-flex items-center gap-1'>
                      <div className='w-2 h-2 bg-blue-600 dark:bg-blue-400 rounded-full'></div>
                      <strong className='text-blue-700 dark:text-blue-300'>blue dots</strong>
                    </span>{' '}
                    between them.
                  </p>
                </div>
              </div>

              {/* Step 3 */}
              <div className='flex gap-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700'>
                <div className='flex items-center justify-center w-6 h-6 bg-green-100 dark:bg-green-900 rounded-full shrink-0 mt-0.5'>
                  <div className='text-xs font-medium text-green-600 dark:text-green-300'>3</div>
                </div>
                <div>
                  <h4 className='font-medium text-gray-900 dark:text-gray-100 mb-1'>
                    Configure Data Flow
                  </h4>
                  <p className='text-gray-600 dark:text-gray-300 text-sm leading-relaxed'>
                    Some phases produce data that others need. You can define this flow by linking
                    <strong> outputs</strong> from earlier phases to the <strong>inputs</strong> of
                    later ones.
                    <br className='my-1' />
                    Use{' '}
                    <span className='inline-flex items-center gap-1'>
                      <div className='w-2 h-2 bg-green-600 dark:bg-green-400 rounded-full'></div>
                      <strong className='text-green-700 dark:text-green-300'>green dots</strong>
                    </span>{' '}
                    for participant-level data and{' '}
                    <span className='inline-flex items-center gap-1'>
                      <div className='w-2 h-2 bg-purple-600 dark:bg-purple-400 rounded-full'></div>
                      <strong className='text-purple-700 dark:text-purple-300'>purple dots</strong>
                    </span>{' '}
                    for phase-level data.
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Rules */}
          <div className='p-4 rounded-lg border border-amber-200 dark:border-amber-700'>
            <h3 className='font-medium text-amber-900 dark:text-amber-100 mb-2 flex items-center gap-2'>
              <ArrowRight className='h-4 w-4 text-amber-700 dark:text-amber-300' />
              Connection Rules
            </h3>
            <p className='text-amber-800 dark:text-amber-200 text-sm leading-relaxed'>
              The configurator follows a <strong>linear and forward-only rule</strong>. This means:
              <ul className='list-disc list-inside mt-2 space-y-1'>
                <li>
                  You must define a single, straight path for how students progress through the
                  phases.
                </li>
                <li>Loops or branching (parallel) paths are not allowed.</li>
                <li>
                  Data connections can only go forward—only to phases that come later in the student
                  flow.
                </li>
              </ul>
            </p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
