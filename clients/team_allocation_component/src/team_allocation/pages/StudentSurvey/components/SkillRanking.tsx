import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  RadioGroup,
  RadioGroupItem,
  Separator,
} from '@tumaet/prompt-ui-components'
import { Star } from 'lucide-react'
import { Skill } from '../../../interfaces/skill'
import { SkillLevel } from '../../../interfaces/skillResponse'
import React from 'react'

interface SkillRankingProps {
  skills: Skill[]
  skillRatings: Record<string, SkillLevel | undefined>
  setSkillRatings: React.Dispatch<React.SetStateAction<Record<string, SkillLevel | undefined>>>
  disabled: boolean
}

const skillLevelOptions = [
  { label: 'Very Bad', value: SkillLevel.VERY_BAD },
  { label: 'Bad', value: SkillLevel.BAD },
  { label: 'Ok', value: SkillLevel.OK },
  { label: 'Good', value: SkillLevel.GOOD },
  { label: 'Very Good', value: SkillLevel.VERY_GOOD },
]

export const SkillRanking = ({
  skills,
  skillRatings,
  setSkillRatings,
  disabled,
}: SkillRankingProps) => {
  const handleSkillRatingChange = (skillID: string, skillLevel: SkillLevel) => {
    setSkillRatings((prev) => ({ ...prev, [skillID]: skillLevel }))
  }

  return (
    <Card>
      <CardHeader className='bg-muted/50'>
        <div className='flex items-center justify-between'>
          <div>
            <CardTitle className='flex items-center gap-2'>
              <Star className='h-5 w-5 text-primary' />
              Personal Proficiency Level
            </CardTitle>
            <CardDescription>
              Rate your personal proficiency level in the following areas:
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className='pt-6'>
        <div className='space-y-6'>
          {skills.map((skill) => (
            <div key={skill.id} className='space-y-2'>
              <div className='flex items-center justify-between'>
                <span className='font-medium'>{skill.name}</span>
              </div>

              <RadioGroup
                value={skillRatings[skill.id] || ''}
                onValueChange={(val) => handleSkillRatingChange(skill.id, val as SkillLevel)}
                className='flex justify-between gap-2 pt-1'
                disabled={disabled}
              >
                {skillLevelOptions.map(({ label, value }) => (
                  <div key={value} className='flex flex-col items-center gap-1'>
                    <RadioGroupItem value={value} id={`${skill.id}-${value}`} className='h-5 w-5' />
                    <label
                      htmlFor={`${skill.id}-${value}`}
                      className='text-xs text-muted-foreground cursor-pointer'
                    >
                      {label}
                    </label>
                  </div>
                ))}
              </RadioGroup>

              <Separator className='mt-4' />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
