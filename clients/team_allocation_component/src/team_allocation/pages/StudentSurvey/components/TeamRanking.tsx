import {
  Badge,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { ClipboardList, GripVertical } from 'lucide-react'
import type { Team } from '@tumaet/prompt-shared-state'
import { DragDropContext, Droppable, Draggable, type DropResult } from '@hello-pangea/dnd'

interface TeamRankingProps {
  teamRanking: string[]
  teams: Team[]
  setTeamRanking: (teamRanking: string[]) => void
  disabled: boolean
  preferenceMode: 'teams' | 'fields'
}

export const TeamRanking = ({
  teamRanking,
  teams,
  setTeamRanking,
  disabled,
  preferenceMode,
}: TeamRankingProps) => {
  const handleDragEnd = (result: DropResult) => {
    if (!result.destination) return
    const newOrder = Array.from(teamRanking)
    const [removed] = newOrder.splice(result.source.index, 1)
    newOrder.splice(result.destination.index, 0, removed)
    setTeamRanking(newOrder)
  }
  const isFieldMode = preferenceMode === 'fields'

  return (
    <Card>
      <CardHeader className='bg-muted/50'>
        <div className='flex items-center justify-between'>
          <div>
            <CardTitle className='flex items-center gap-2'>
              <ClipboardList className='h-5 w-5 text-primary' />
              {isFieldMode ? 'Field Preferences' : 'Team Preferences'}
            </CardTitle>
            <CardDescription>
              {isFieldMode
                ? 'Rank the fields by your preference (1st is most preferred). Drag and drop to reorder.'
                : 'Rank the teams by your preference (1st is most preferred). Drag and drop to reorder.'}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className='pt-6'>
        <DragDropContext onDragEnd={handleDragEnd}>
          <Droppable droppableId='teamRanking' isDropDisabled={disabled}>
            {(provided) => (
              <div className='space-y-4' {...provided.droppableProps} ref={provided.innerRef}>
                {teamRanking.map((teamID, index) => {
                  const team = teams.find((t) => t.id === teamID)
                  if (!team) return null
                  return (
                    <Draggable key={team.id} draggableId={team.id} index={index}>
                      {(prov, snapshot) => (
                        <div
                          ref={prov.innerRef}
                          {...prov.draggableProps}
                          {...prov.dragHandleProps}
                          className={`flex items-center justify-between p-3 border rounded-lg bg-card transition-colors ${
                            snapshot.isDragging ? 'bg-accent/10' : 'hover:bg-accent/5'
                          }`}
                        >
                          <div className='flex items-center gap-3'>
                            <Badge
                              variant='outline'
                              className='w-8 h-8 rounded-full flex items-center justify-center p-0'
                            >
                              {index + 1}
                            </Badge>
                            <span className='font-medium ml-2'>{team.name}</span>
                          </div>
                          <GripVertical
                            className='h-5 w-5 text-muted-foreground cursor-grab'
                            aria-label='Drag handle'
                          />
                        </div>
                      )}
                    </Draggable>
                  )
                })}
                {provided.placeholder}
              </div>
            )}
          </Droppable>
        </DragDropContext>
      </CardContent>
    </Card>
  )
}
