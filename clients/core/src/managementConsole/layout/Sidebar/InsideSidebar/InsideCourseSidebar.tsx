import { CourseConfiguratorSidebar } from '@managementConsole/PhaseMapping/ExternalSidebars/CourseConfiguratorSidebar'
import { CourseSettingsSidebar } from '@managementConsole/PhaseMapping/ExternalSidebars/CourseSettingsSidebar'
import { CourseUserManagementSidebar } from '@managementConsole/PhaseMapping/ExternalSidebars/CourseUserManagementSidebar'
import { PhaseSidebarMapping } from '@managementConsole/PhaseMapping/PhaseSidebarMapping'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
} from '@tumaet/prompt-ui-components'
import { Gauge } from 'lucide-react'
import { Suspense, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { DisabledSidebarMenuItem } from './components/DisabledSidebarMenuItem'
import { InsideSidebarMenuItem } from './components/InsideSidebarMenuItem'

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
    <SidebarMenu>
      <SidebarGroup>
        <SidebarGroupContent>
          <InsideSidebarMenuItem goToPath={rootPath} icon={<Gauge />} title='Overview' />
          <CourseConfiguratorSidebar rootPath={rootPath} title='Course Configurator' />
          <CourseSettingsSidebar rootPath={rootPath} title='Settings' />
          <CourseUserManagementSidebar rootPath={rootPath} title='User Management' />
        </SidebarGroupContent>
      </SidebarGroup>

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
                    rootPath={`${rootPath}/${phase.id}`}
                    title={phase.name}
                    coursePhaseID={phase.id}
                  />
                </Suspense>
              )
            } else {
              return <DisabledSidebarMenuItem key={phase.id} title={`Unknown ${phase.name}`} />
            }
          })}

          {/*
            Empty state shown when no phase menu item is visible to the current user.
            This covers a course with no phases and a course whose phases are all
            hidden by permissions (e.g. a student whose phases have no student-facing
            sidebar). The sibling selector hides this block as soon as any real phase
            menu item renders above it, so we rely on what is actually rendered rather
            than re-deriving per-phase visibility (which depends on async participation
            data and remote sidebar configs).
          */}
          <p className='px-2 text-xs text-muted-foreground [[data-sidebar=menu-item]~&]:hidden'>
            No course phases yet.
          </p>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarMenu>
  )
}
