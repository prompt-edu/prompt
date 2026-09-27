import { useApplicationStore } from '@core/managementConsole/applicationAdministration/zustand/useApplicationStore'
import { useStudentStore } from '@core/managementConsole/shared/store/student.store'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Breadcrumb,
  BreadcrumbEllipsis,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  getStudentName,
} from '@tumaet/prompt-ui-components'
import React, { useLayoutEffect, useMemo, useRef, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { collapseBreadcrumbs } from './collapseBreadcrumbs'

interface BreadcrumbProps {
  title: string
  path: string
}

const capitalizeFirstLetter = (string: string) => {
  return string.charAt(0).toUpperCase() + string.slice(1)
}

const SEGMENT_LABELS: Record<string, string> = {
  'tease-config': 'TEASE Configuration',
}

const segmentLabel = (segment: string) => SEGMENT_LABELS[segment] ?? capitalizeFirstLetter(segment)

export const Breadcrumbs: React.FC = () => {
  const location = useLocation()
  const navigate = useNavigate()
  const { courses } = useCourseStore()
  const { studentsById } = useStudentStore()
  const { participations } = useApplicationStore()

  const breadcrumbList = useMemo(() => {
    const pathSegments = location.pathname.split('/').filter(Boolean)
    const breadcrumbs: BreadcrumbProps[] = []

    if (pathSegments[0] === 'management') {
      if (pathSegments[1] === 'courses') {
        breadcrumbs.push({ title: 'Courses', path: '/management/courses' })
        pathSegments.slice(2).forEach((segment, index) => {
          breadcrumbs.push({
            title: segment.toUpperCase(),
            path: `/management/courses/${pathSegments.slice(2, index + 3).join('/')}`,
          })
        })
      } else if (pathSegments[1] === 'course-templates') {
        breadcrumbs.push({ title: 'Template Courses', path: '/management/course-templates' })
      } else if (pathSegments[1] === 'course-archive') {
        breadcrumbs.push({ title: 'Archived Courses', path: '/management/course-archive' })
      } else if (pathSegments[1] === 'privacy') {
        breadcrumbs.push({ title: 'Privacy', path: '/management/privacy' })
        if (pathSegments[2] === 'data-export') {
          breadcrumbs.push({ title: 'Data Export', path: '/management/privacy/data-export' })
        } else if (pathSegments[2] === 'data-deletion') {
          breadcrumbs.push({ title: 'Data Deletion', path: '/management/privacy/data-deletion' })
        }
      } else if (pathSegments[1] === 'students') {
        breadcrumbs.push({ title: 'Students', path: '/management/students' })
        if (pathSegments.length > 2) {
          if (studentsById[pathSegments[2]]) {
            const s = studentsById[pathSegments[2]]
            breadcrumbs.push({
              title: getStudentName(s),
              path: `/management/students/${pathSegments[2]}`,
            })
          } else {
            breadcrumbs.push({ title: 'Student', path: `/management/students/${pathSegments[2]}` })
          }
        }
      } else if (pathSegments[1] === 'course' && pathSegments.length >= 3) {
        const courseId = pathSegments[2]
        const course = courses.find((c) => c.id === courseId)
        if (course) {
          breadcrumbs.push({ title: course.name, path: `/management/course/${courseId}` })

          if (pathSegments.length >= 3 && pathSegments[3] === 'configurator') {
            breadcrumbs.push({
              title: 'Course Configurator',
              path: `/management/course/${courseId}/configurator`,
            })
          } else if (pathSegments.length >= 3) {
            const phaseId = pathSegments[3]
            const phase = course.coursePhases.find((p) => p.id === phaseId)
            if (phase) {
              breadcrumbs.push({
                title: phase.name,
                path: `/management/course/${courseId}/${phaseId}`,
              })
            }
            pathSegments.slice(4).forEach((segment, index) => {
              // we assume that longer items are courseParticipationIDs
              if (segment.length < 20) {
                breadcrumbs.push({
                  title: segmentLabel(segment),
                  path: `/management/course/${courseId}/${phaseId}/${pathSegments.slice(4, index + 5).join('/')}`,
                })
              } else {
                // This is likely a courseParticipationID (long UUID)
                const participation = participations.find(
                  (p) => p.courseParticipationID === segment,
                )
                const fullName = [
                  participation?.student?.firstName ?? '',
                  participation?.student?.lastName ?? '',
                ]
                  .filter(Boolean)
                  .join(' ')
                const title = fullName || 'Participant'
                breadcrumbs.push({
                  title,
                  path: `/management/course/${courseId}/${phaseId}/${pathSegments.slice(4, index + 5).join('/')}`,
                })
              }
            })
          }
        }
      }
    }

    return breadcrumbs
  }, [location.pathname, courses, studentsById, participations])

  const containerRef = useRef<HTMLElement>(null)
  const fullTrailRef = useRef<HTMLOListElement>(null)
  const [isCollapsed, setIsCollapsed] = useState(false)

  // Measures an invisible copy of the full trail against the space the header leaves, so the
  // trail collapses as soon as it would wrap and expands again once it fits.
  useLayoutEffect(() => {
    const container = containerRef.current
    const fullTrail = fullTrailRef.current
    if (!container || !fullTrail) return

    const update = () => setIsCollapsed(fullTrail.scrollWidth > container.clientWidth)
    update()
    const observer = new ResizeObserver(update)
    observer.observe(container)
    observer.observe(fullTrail)
    return () => observer.disconnect()
  }, [breadcrumbList])

  if (breadcrumbList.length === 0) {
    return null
  }

  const renderCrumb = (crumb: BreadcrumbProps, isLast: boolean) => (
    <BreadcrumbItem className={isLast ? 'min-w-0' : 'min-w-0 max-w-48 shrink-0'}>
      {isLast ? (
        <BreadcrumbPage className='block truncate' title={crumb.title}>
          {crumb.title}
        </BreadcrumbPage>
      ) : (
        <BreadcrumbLink
          className='block truncate'
          title={crumb.title}
          style={{ cursor: 'pointer' }}
          onClick={() => navigate(crumb.path)}
        >
          {crumb.title}
        </BreadcrumbLink>
      )}
    </BreadcrumbItem>
  )

  const renderFullTrail = () =>
    breadcrumbList.map((crumb, index) => (
      <React.Fragment key={crumb.path}>
        {index > 0 && <BreadcrumbSeparator />}
        {renderCrumb(crumb, index === breadcrumbList.length - 1)}
      </React.Fragment>
    ))

  const { first, hidden, last } = collapseBreadcrumbs(breadcrumbList)

  return (
    <Breadcrumb ref={containerRef} className='relative min-w-0 flex-1'>
      <BreadcrumbList
        ref={fullTrailRef}
        aria-hidden='true'
        className='invisible absolute w-max flex-nowrap'
      >
        {renderFullTrail()}
      </BreadcrumbList>
      <BreadcrumbList className='flex-nowrap'>
        {isCollapsed && first && last ? (
          <>
            {renderCrumb(first, false)}
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <DropdownMenu>
                <DropdownMenuTrigger
                  className='flex items-center'
                  aria-label='Show hidden breadcrumbs'
                >
                  <BreadcrumbEllipsis className='h-4 w-4' />
                </DropdownMenuTrigger>
                <DropdownMenuContent align='start'>
                  {hidden.map((crumb) => (
                    <DropdownMenuItem key={crumb.path} onClick={() => navigate(crumb.path)}>
                      {crumb.title}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            {renderCrumb(last, true)}
          </>
        ) : (
          renderFullTrail()
        )}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
