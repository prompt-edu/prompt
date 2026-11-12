import { SkillProficiency } from '../../../interfaces/tease/skillProficiency'

export const getLevelConfig = (level: SkillProficiency, unknown?: boolean) => {
  if (unknown) {
    return {
      title: 'Unknown',
      textColor: 'text-gray-700',
      selectedBg: 'bg-gray-100',
      icon: '⚪',
      border: 'border-gray-300',
    }
  }

  switch (level) {
    case SkillProficiency.Novice:
      return {
        title: 'Novice',
        textColor: 'text-red-700',
        selectedBg: 'bg-red-100',
        icon: '🔴',
        border: 'border-red-300',
      }
    case SkillProficiency.Intermediate:
      return {
        title: 'Intermediate',
        textColor: 'text-yellow-700',
        selectedBg: 'bg-yellow-100',
        icon: '🟡',
        border: 'border-yellow-300',
      }
    case SkillProficiency.Advanced:
      return {
        title: 'Advanced',
        textColor: 'text-green-700',
        selectedBg: 'bg-green-100',
        icon: '🟢',
        border: 'border-green-300',
      }
    case SkillProficiency.Expert:
      return {
        title: 'Expert',
        textColor: 'text-blue-700',
        selectedBg: 'bg-blue-100',
        icon: '🔵',
        border: 'border-blue-300',
      }
  }
}
