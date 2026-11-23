import { courseOrgID } from '@core/managementConsole/shared/utils/orgRoles'
import { useOrgs, useOwnOrgs, useSetCourseOrg } from '@core/network/hooks/useOrgs'
import { type Course, Role, useAuthStore } from '@tumaet/prompt-shared-state'
import {
  Button,
  CardContent,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  SettingsCard,
} from '@tumaet/prompt-ui-components'
import { Building2, Loader2, Save } from 'lucide-react'
import { useState } from 'react'

interface CourseOrgSettingsProps {
  course: Course
}

const NO_ORG = 'none'

// The course's org. Only PROMPT admins assign it, so nobody claims a course for their org.
export const CourseOrgSettings = ({ course }: CourseOrgSettingsProps) => {
  const { permissions } = useAuthStore()
  const isPromptAdmin = permissions.includes(Role.PROMPT_ADMIN)
  const { data: allOrgs = [] } = useOrgs(isPromptAdmin)
  const { data: ownOrgs = [] } = useOwnOrgs()
  const setCourseOrg = useSetCourseOrg(course.id)

  const savedOrgID = courseOrgID(course) ?? NO_ORG
  const [orgID, setOrgID] = useState(savedOrgID)
  const orgName = [...allOrgs, ...ownOrgs].find((org) => org.id === savedOrgID)?.name

  if (!isPromptAdmin) {
    if (savedOrgID === NO_ORG) {
      return null
    }
    return (
      <SettingsCard
        icon={<Building2 className='w-5 h-5' />}
        title='Org'
        description='The Admins of the org manage all of its courses. PROMPT administrators assign courses to orgs.'
      >
        <CardContent>
          <p>{orgName ?? 'This course belongs to an org you are not a member of.'}</p>
        </CardContent>
      </SettingsCard>
    )
  }

  return (
    <SettingsCard
      icon={<Building2 className='w-5 h-5' />}
      title='Org'
      description='The Admins of the org manage all of its courses.'
    >
      <CardContent className='flex flex-wrap items-center gap-4'>
        <Select value={orgID} onValueChange={setOrgID}>
          <SelectTrigger className='w-72' aria-label='Org of the course'>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={NO_ORG}>No org</SelectItem>
            {allOrgs.map((org) => (
              <SelectItem key={org.id} value={org.id}>
                {org.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          disabled={orgID === savedOrgID || setCourseOrg.isPending}
          onClick={() => setCourseOrg.mutate(orgID === NO_ORG ? null : orgID)}
        >
          {setCourseOrg.isPending ? (
            <Loader2 className='w-4 h-4 animate-spin' />
          ) : (
            <Save className='w-4 h-4' />
          )}
          Save
        </Button>
      </CardContent>
    </SettingsCard>
  )
}
