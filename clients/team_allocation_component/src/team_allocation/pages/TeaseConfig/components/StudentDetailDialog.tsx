import { useQuery } from '@tanstack/react-query'
import type { Team } from '@tumaet/prompt-shared-state'
import {
  Badge,
  Card,
  CardContent,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  ErrorPage,
  Separator,
} from '@tumaet/prompt-ui-components'
import { Award, BookOpen, Loader2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import type { TeaseStudent } from '../../../interfaces/tease/student'
import { getAllSkills } from '../../../network/queries/getAllSkills'
import { getAllTeams } from '../../../network/queries/getAllTeams'
import { getLevelConfig } from './ProficiencyBadge'

interface StudentDetailDialogProps {
  student: TeaseStudent | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function StudentDetailDialog({ student, open, onOpenChange }: StudentDetailDialogProps) {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: teams,
    isPending: isTeamsPending,
    isError: isTeamsError,
    refetch: refetchTeams,
  } = useQuery<Team[]>({
    queryKey: ['tease_teams', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
  })

  const {
    data: skills,
    isPending: isSkillsPending,
    isError: isSkillsError,
    refetch: refetchSkills,
  } = useQuery({
    queryKey: ['tease_skills', phaseId],
    queryFn: () => getAllSkills(phaseId ?? ''),
  })

  function getTeamNameById(teamId: string): string {
    const team = teams?.find((t) => t.id === teamId)
    return team ? team.name : teamId
  }

  function getSkillNameById(skillId: string): string {
    const skill = skills?.find((s) => s.id === skillId)
    return skill ? skill.name : skillId
  }

  if (!student) return null

  if (isTeamsPending || isSkillsPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  if (isTeamsError || isSkillsError) {
    return (
      <ErrorPage
        onRetry={() => {
          refetchTeams()
          refetchSkills()
        }}
      />
    )
  }

  // Sort project preferences by priority (assuming lower number means higher priority)
  const sortedPreferences = [...(student.projectPreferences || [])].sort(
    (a, b) => (a.priority || 0) - (b.priority || 0),
  )

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className=''>
        <div className='flex flex-col h-full sm:max-w-[650px] max-h-[90vh] overflow-hidden'>
          <DialogHeader className='bg-background z-5 pb-2'>
            <DialogTitle className='text-xl flex items-center gap-2'>
              <span className='font-semibold'>
                {student.firstName} {student.lastName}
              </span>
              <Badge variant='outline' className='ml-2'>
                {student.email}
              </Badge>
            </DialogTitle>
            <DialogDescription>Skills and Project Preferences Details</DialogDescription>
          </DialogHeader>

          <Separator className='mt-2' />

          <div className='flex-1 pr-1 overflow-y-auto'>
            <div className='space-y-6 pr-2 mt-4'>
              <div>
                <div className='flex items-center gap-2 mb-3'>
                  <Award className='h-5 w-5 text-primary' />
                  <h3 className='text-lg font-medium'>Skills & Proficiency</h3>
                </div>

                {student.skills && student.skills.length > 0 ? (
                  <div className='space-y-4'>
                    {student.skills.map((skill, index) => (
                      <Card key={index} className='overflow-hidden'>
                        <CardContent className='p-3'>
                          <div className='flex flex-col sm:flex-row sm:items-center gap-2'>
                            <div className='font-medium'>{getSkillNameById(skill.id)}</div>
                            <div className='w-auto'>
                              <Badge
                                className={`
                                  ${getLevelConfig(skill.proficiency).textColor} 
                                  ${getLevelConfig(skill.proficiency).selectedBg} 
                                  hover:${getLevelConfig(skill.proficiency).selectedBg}
                                `}
                              >
                                {getLevelConfig(skill.proficiency).title}
                              </Badge>
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                ) : (
                  <p className='text-sm text-muted-foreground italic'>No skills provided</p>
                )}
              </div>

              <div>
                <div className='flex items-center gap-2 mb-3'>
                  <BookOpen className='h-5 w-5 text-primary' />
                  <h3 className='text-lg font-medium'>Project Preferences</h3>
                </div>

                {sortedPreferences.length > 0 ? (
                  <div className='space-y-4'>
                    {sortedPreferences.map((preference) => (
                      <Card key={preference.projectId} className='overflow-hidden'>
                        <CardContent className='p-3'>
                          <div className='flex justify-between items-center'>
                            <div className='flex items-center gap-2'>
                              <div className='bg-muted w-7 h-7 rounded-full flex items-center justify-center font-medium text-sm'>
                                {(preference.priority ?? 0) + 1}
                              </div>
                              <div className='font-medium'>
                                {getTeamNameById(preference.projectId)}
                              </div>
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                ) : (
                  <p className='text-sm text-muted-foreground italic'>
                    No project preferences submitted
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
