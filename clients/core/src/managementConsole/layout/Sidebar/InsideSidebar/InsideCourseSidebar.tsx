import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
} from '@tumaet/prompt-ui-components'
import { Gauge } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { InsideSidebarMenuItem } from './components/InsideSidebarMenuItem'
import { Suspense, useMemo } from 'react'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import { DisabledSidebarMenuItem } from './components/DisabledSidebarMenuItem'
import { PhaseSidebarMapping } from '@managementConsole/PhaseMapping/PhaseSidebarMapping'
import { CourseConfiguratorSidebar } from '@managementConsole/PhaseMapping/ExternalSidebars/CourseConfiguratorSidebar'
import { CourseSettingsSidebar } from '@managementConsole/PhaseMapping/ExternalSidebars/CourseSettingsSidebar'
import { CourseUserManagementSidebar } from '@managementConsole/PhaseMapping/ExternalSidebars/CourseUserManagementSidebar'

export const InsideCourseSidebar = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const { courses } = useCourseStore()

  const rootPath = `/management/course/${courseId}`

  const { sortedPhases } = useMemo(() => {
    const activeCourse = courses.find((c) => c.id === courseId)
    const phasesSorted =
      activeCourse?.coursePhases
        .filter((phase) => phase.sequenceOrder !== -1) // filter out the ones without sequence order
        .sort((a, b) => a.sequenceOrder - b.sequenceOrder) || []
    return { sortedPhases: phasesSorted }
  }, [courseId, courses])

  return (
    <SidebarMenu className='space-y-4'>
      <SidebarGroup>
        <SidebarGroupContent>
          <InsideSidebarMenuItem goToPath={rootPath} icon={<Gauge />} title='Overview' />
          <CourseConfiguratorSidebar rootPath={rootPath} title='Course Configurator' />
          <CourseSettingsSidebar rootPath={rootPath} title='Settings' />
          <CourseUserManagementSidebar rootPath={rootPath} title='User Management' />
        </SidebarGroupContent>
      </SidebarGroup>

      {sortedPhases.length > 0 && (
        <SidebarGroup>
          <SidebarGroupLabel>Course Phases</SidebarGroupLabel>
          <SidebarGroupContent>
            {sortedPhases.map((phase) => {
              if (phase.coursePhaseType in PhaseSidebarMapping) {
                const PhaseComponent = PhaseSidebarMapping[phase.coursePhaseType]
                return (
                  <Suspense
                    key={phase.id}
                    fallback={<DisabledSidebarMenuItem key={phase.id} title={'Loading...'} />}
                  >
                    <PhaseComponent
                      rootPath={rootPath + '/' + phase.id}
                      title={phase.name}
                      coursePhaseID={phase.id}
                    />
                  </Suspense>
                )
              } else {
                return <DisabledSidebarMenuItem key={phase.id} title={'Unknown ' + phase.name} />
              }
            })}
          </SidebarGroupContent>
        </SidebarGroup>
      )}
    </SidebarMenu>
  )
}
