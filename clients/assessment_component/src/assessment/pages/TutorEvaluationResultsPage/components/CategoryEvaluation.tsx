import { useState } from 'react'
import { ChevronDown, ChevronRight } from 'lucide-react'

import { AssessmentType } from '../../../interfaces/assessmentType'
import { CategoryWithCompetencies } from '../../../interfaces/category'
import { Evaluation } from '../../../interfaces/evaluation'
import { mapNumberToScoreLevel, mapScoreLevelToNumber } from '@tumaet/prompt-shared-state'

import { useTeamStore } from '../../../zustand/useTeamStore'

import { CompetencyHeader } from '../../components/CompetencyHeader'
import { ScoreLevelSelector } from '../../components/ScoreLevelSelector'
import { StudentScoreBadge } from '../../components/badges'

import { getLevelConfig } from '@tumaet/prompt-ui-components'
import { getWeightedScoreLevel } from '../../utils/getWeightedScoreLevel'

interface CategoryEvaluationProps {
  category: CategoryWithCompetencies
  evaluations: Evaluation[]
}

export const CategoryEvaluation = ({ category, evaluations }: CategoryEvaluationProps) => {
  const { teams } = useTeamStore()

  const getTeamMemberName = (authorCourseParticipationID: string) => {
    for (const team of teams) {
      const member = team.members.find((m) => m.id === authorCourseParticipationID)
      if (member) {
        return `${member.firstName} ${member.lastName}`
      }
    }
    return 'Unknown member'
  }

  const [isExpanded, setIsExpanded] = useState(true)

  const toggleExpand = () => {
    setIsExpanded(!isExpanded)
  }

  const categoryScore = getWeightedScoreLevel(evaluations, [category])

  return (
    <div key={category.id} className='mb-6'>
      <div className='flex grow items-center justify-center mb-4 gap-1'>
        <div className='flex items-center justify-center w-full'>
          <button
            onClick={toggleExpand}
            className='p-1 mr-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-xs focus:outline-hidden'
            aria-expanded={isExpanded}
            aria-controls={`content-${category.id}`}
          >
            {isExpanded ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
          </button>

          <h2 className='text-xl font-semibold tracking-tight grow'>{category.name}</h2>
        </div>
        <div className='flex items-center justify-center gap-1'>
          {evaluations.length > 0 && (
            <StudentScoreBadge
              scoreLevel={mapNumberToScoreLevel(categoryScore)}
              scoreNumeric={categoryScore}
            />
          )}
        </div>
      </div>

      {isExpanded && (
        <div id={`content-${category.id}`} className='pt-4 pb-2 space-y-5 border-t mt-2'>
          {category.competencies.map((competency) => {
            const competencyEvaluations = evaluations.filter(
              (e) => e.competencyID === competency.id,
            )
            const avgNumeric =
              competencyEvaluations.length > 0
                ? competencyEvaluations.reduce(
                    (sum, evaluation) => sum + mapScoreLevelToNumber(evaluation.scoreLevel),
                    0,
                  ) / competencyEvaluations.length
                : undefined
            const evaluationAverageScoreLevel =
              avgNumeric === undefined ? undefined : mapNumberToScoreLevel(avgNumeric)
            const studentNames = [
              () => (
                <span>
                  {competencyEvaluations.map((evaluation) => (
                    <div key={evaluation.id} className='flex items-center gap-2'>
                      <span
                        className={`px-2 py-1 rounded-sm text-xs font-medium cursor-help ${
                          getLevelConfig(evaluation.scoreLevel).selectedBg
                        }`}
                      >
                        {getLevelConfig(evaluation.scoreLevel).title}
                      </span>
                      <span>{getTeamMemberName(evaluation.authorCourseParticipationID)}</span>
                    </div>
                  ))}
                </span>
              ),
            ]

            return (
              <div key={competency.id} className='mb-6 last:mb-0'>
                <div className='grid grid-cols-1 lg:grid-cols-2 gap-4 items-start p-4 border rounded-md relative'>
                  <CompetencyHeader
                    className='lg:col-span-2'
                    assessmentType={AssessmentType.TUTOR}
                    competency={competency}
                    completed={true}
                    onResetClick={() => {}}
                  />

                  <ScoreLevelSelector
                    className='lg:col-span-2 grid grid-cols-1 lg:grid-cols-5 gap-1'
                    assessmentType={AssessmentType.TUTOR}
                    competency={competency}
                    onScoreChange={() => {}}
                    completed={false}
                    selectedScore={evaluationAverageScoreLevel}
                    peerEvaluationCompetency={competency}
                    peerEvaluationScoreLevel={evaluationAverageScoreLevel}
                    peerEvaluationStudentAnswers={studentNames}
                  />
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
