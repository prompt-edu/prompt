import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@tumaet/prompt-ui-components'
import { Bar, BarChart, Cell, LabelList, XAxis, YAxis } from 'recharts'

import { GRADE_CONFIG } from '../../utils/gradeConfig'

interface GradeDataPoint {
  dataKey: string
  count: number
  total: number
}

interface BarChartWithGradesProps {
  data: GradeDataPoint[]
}

const getColorForGrade = (grade: string): string => {
  return GRADE_CONFIG[grade as keyof typeof GRADE_CONFIG]?.color || '#d4d4d8'
}

export const BarChartWithGrades = ({ data }: BarChartWithGradesProps) => {
  return (
    <ChartContainer config={GRADE_CONFIG} className='mx-auto w-full h-[280px]'>
      <BarChart data={data} margin={{ top: 20, right: 20, bottom: 20, left: 20 }}>
        <XAxis
          dataKey='dataKey'
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 11 }}
          interval={0}
          height={60}
          angle={-45}
          textAnchor='end'
        />
        <YAxis hide />
        <ChartTooltip
          cursor={false}
          content={<ChartTooltipContent />}
          labelFormatter={(label) => `Grade: ${label}`}
          formatter={(value) => [value, ' Students']}
        />
        <Bar dataKey='count' radius={[4, 4, 0, 0]}>
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={getColorForGrade(entry.dataKey)} />
          ))}
          <LabelList
            dataKey='count'
            position='top'
            offset={10}
            className='fill-foreground'
            fontSize={12}
          />
        </Bar>
      </BarChart>
    </ChartContainer>
  )
}
