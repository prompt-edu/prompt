import { useApplicationStore } from '@core/managementConsole/applicationAdministration/zustand/useApplicationStore'
import { useStudentStore } from '@core/managementConsole/shared/store/student.store'
import { coreKeys } from '@core/network/cache'
import { skipToken, useQuery } from '@tanstack/react-query'
import {
  type CoursePhaseParticipationsWithResolution,
  useCourseStore,
} from '@tumaet/prompt-shared-state'
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
import { getParticipantName } from './getParticipantName'

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
  const [, section, , phaseIdInPath] = location.pathname.split('/').filter(Boolean)
  const { data: phaseParticipations } = useQuery<CoursePhaseParticipationsWithResolution>({
    queryKey: coreKeys.coursePhases.participants(section === 'course' ? phaseIdInPath : undefined),
    queryFn: skipToken,
  })

  const breadcrumbList = useMemo(() => {
    const pathSegments = location.pathname.split('/').filter(Boolean)
    const breadcrumbs: BreadcrumbProps[] = []
    const knownParticipations = [...participations, ...(phaseParticipations?.participations ?? [])]

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
                // This is likely a courseParticipationID or a student ID (long UUID)
                breadcrumbs.push({
                  title: getParticipantName(segment, knownParticipations) ?? 'Participant',
                  path: `/management/course/${courseId}/${phaseId}/${pathSegments.slice(4, index + 5).join('/')}`,
                })
              }
            })
          }
        }
      }
    }

    return breadcrumbs
  }, [location.pathname, courses, studentsById, participations, phaseParticipations])

  const { first, hidden, last } = collapseBreadcrumbs(breadcrumbList)
  const canCollapse = hidden.length > 0

  const containerRef = useRef<HTMLElement>(null)
  const listRef = useRef<HTMLOListElement>(null)
  const lastLabelRef = useRef<HTMLSpanElement>(null)
  const fullWidthRef = useRef(0)
  const [isCollapsed, setIsCollapsed] = useState(false)

  // A new trail has not been measured yet. Forgetting the old width makes the measuring effect
  // show it in full, measure it, and collapse it again if needed, all before paint.
  useLayoutEffect(() => {
    fullWidthRef.current = 0
  }, [breadcrumbList])

  // While the full trail is shown, remember the width it needs (the last crumb may already be
  // truncated, so its untruncated width counts). Collapse whenever that width exceeds the space
  // the header leaves, and expand again once it fits. Runs before paint, so nothing flickers.
  useLayoutEffect(() => {
    const container = containerRef.current
    const list = listRef.current
    if (!container || !list) return

    const update = () => {
      if (!isCollapsed) {
        const lastLabel = lastLabelRef.current
        const lastLabelOverflow = lastLabel ? lastLabel.scrollWidth - lastLabel.clientWidth : 0
        fullWidthRef.current = list.scrollWidth + lastLabelOverflow
      }
      setIsCollapsed(canCollapse && fullWidthRef.current > container.clientWidth)
    }
    update()
    const observer = new ResizeObserver(update)
    observer.observe(container)
    return () => observer.disconnect()
  }, [isCollapsed, canCollapse, breadcrumbList])

  if (breadcrumbList.length === 0) {
    return null
  }

  // The last crumb truncates first. Earlier crumbs are capped only in the collapsed form; in the
  // full form they keep their real width so it can be measured, unless the trail cannot collapse
  // at all, in which case they shrink and truncate instead.
  const earlierCrumbClassName = isCollapsed
    ? 'min-w-0 max-w-48 shrink-0'
    : canCollapse
      ? 'shrink-0'
      : 'min-w-0'

  const renderCrumb = (crumb: BreadcrumbProps, isLast: boolean) => (
    <BreadcrumbItem className={isLast ? 'min-w-0' : earlierCrumbClassName}>
      {isLast ? (
        <BreadcrumbPage ref={lastLabelRef} className='block truncate' title={crumb.title}>
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

  return (
    <Breadcrumb ref={containerRef} className='min-w-0 flex-1'>
      <BreadcrumbList ref={listRef} className='flex-nowrap'>
        {isCollapsed ? (
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
          breadcrumbList.map((crumb, index) => (
            <React.Fragment key={crumb.path}>
              {index > 0 && <BreadcrumbSeparator />}
              {renderCrumb(crumb, index === breadcrumbList.length - 1)}
            </React.Fragment>
          ))
        )}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
