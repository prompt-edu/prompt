import { useMemo, useRef, useState, type ChangeEvent } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, Building2, FileUp, Loader2, Save } from 'lucide-react'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Separator,
} from '@tumaet/prompt-ui-components'
import type { CoursePhaseWithMetaData, UpdateCoursePhase } from '@tumaet/prompt-shared-state'

import { updateCoursePhase } from '@/network/mutations/updateCoursePhase'
import { createTeams } from '../../../network/mutations/createTeams'
import { setAllocationProfile } from '../../../network/mutations/setAllocationProfile'
import type { CompanyImportAnalysis, CompanyRecord } from '../../../interfaces/companyImport'
import { TEAM_ALLOCATION_PROFILE_1000_PLUS } from '../../../interfaces/companyImport'
import {
  analyseCompanyRecords,
  buildCompanyAllocationConfig,
  parseCompanyCsv,
} from '../../../utils/companyCsvImport'

interface CompanyCsvImportProps {
  phaseId: string
  coursePhase: CoursePhaseWithMetaData
}

const countEntries = (entries: Record<string, number>): [string, number][] =>
  Object.entries(entries).sort((first, second) => second[1] - first[1])

export const CompanyCsvImport = ({ phaseId, coursePhase }: CompanyCsvImportProps) => {
  const queryClient = useQueryClient()
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const [companies, setCompanies] = useState<CompanyRecord[]>([])
  const [analysis, setAnalysis] = useState<CompanyImportAnalysis | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const companyNames = useMemo(
    () =>
      Array.from(new Set(companies.map((company) => company.companyName.trim()).filter(Boolean))),
    [companies],
  )

  const fieldNames = useMemo(() => {
    if (!analysis) return []
    return Object.keys(analysis.fieldsOfBusiness).sort((a, b) => a.localeCompare(b))
  }, [analysis])

  const teamSizeConstraints = useMemo(
    () =>
      companies.reduce<Record<string, { lowerBound: number; upperBound: number }>>(
        (constraints, company) => {
          if (company.teamSizeMin === undefined && company.teamSizeMax === undefined) {
            return constraints
          }
          constraints[company.companyName] = {
            lowerBound: company.teamSizeMin ?? 0,
            upperBound: company.teamSizeMax ?? company.teamSizeMin ?? 999,
          }
          return constraints
        },
        {},
      ),
    [companies],
  )

  const saveCompanyConfigMutation = useMutation({
    mutationFn: async (coursePhaseUpdate: UpdateCoursePhase) => {
      if (companyNames.length > 0) {
        await createTeams(
          phaseId,
          companyNames,
          'company_project',
          true,
          Object.keys(teamSizeConstraints).length ? teamSizeConstraints : undefined,
        )
      }
      if (fieldNames.length > 0) {
        await createTeams(phaseId, fieldNames, 'field_bucket', true)
      }
      await updateCoursePhase(coursePhaseUpdate)
      await setAllocationProfile(phaseId, TEAM_ALLOCATION_PROFILE_1000_PLUS)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['course_phase', phaseId] })
      await queryClient.invalidateQueries({ queryKey: ['team_allocation_config', phaseId] })
      await queryClient.invalidateQueries({ queryKey: ['team_allocation_team', phaseId] })
      setSuccessMessage(
        `Applied ${companyNames.length} hidden company projects and ${fieldNames.length} field ranking options.`,
      )
    },
    onError: () => {
      setError('Company allocation configuration could not be saved.')
    },
  })

  const handleFileChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    setError(null)
    setSuccessMessage(null)

    try {
      const csvContent = await file.text()
      const parsedCompanies = parseCompanyCsv(csvContent)
      const nextAnalysis = analyseCompanyRecords(parsedCompanies)
      setCompanies(parsedCompanies)
      setAnalysis(nextAnalysis)
      setSuccessMessage(`Loaded ${parsedCompanies.length} companies from ${file.name}.`)
    } catch (parseError) {
      setCompanies([])
      setAnalysis(null)
      setError(
        parseError instanceof Error ? parseError.message : 'Could not parse the selected CSV file.',
      )
    } finally {
      event.target.value = ''
    }
  }

  const saveCompanyConfig = () => {
    if (!analysis) return

    saveCompanyConfigMutation.mutate({
      id: phaseId,
      restrictedData: {
        ...coursePhase.restrictedData,
        teamAllocationProfile: TEAM_ALLOCATION_PROFILE_1000_PLUS,
        teaseStrategy: TEAM_ALLOCATION_PROFILE_1000_PLUS,
        companyAllocationConfig: buildCompanyAllocationConfig(companies, analysis),
      },
      studentReadableData: coursePhase.studentReadableData,
    })
  }

  return (
    <Card className='w-full shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-start justify-between gap-4'>
          <div>
            <CardTitle className='text-2xl font-bold flex items-center gap-2'>
              <Building2 className='h-5 w-5' />
              Company CSV Import
            </CardTitle>
            <CardDescription className='mt-1.5'>
              Upload companies, extract field alignment options, and save the hidden company project
              configuration for TEASE.
            </CardDescription>
          </div>
          <Badge variant='outline'>{companies.length} companies</Badge>
        </div>
      </CardHeader>

      <Separator />

      <CardContent className='pt-6 space-y-5'>
        <div className='flex flex-wrap gap-2'>
          <input
            ref={fileInputRef}
            type='file'
            accept='.csv,text/csv'
            className='hidden'
            onChange={handleFileChange}
          />
          <Button type='button' className='gap-2' onClick={() => fileInputRef.current?.click()}>
            <FileUp className='h-4 w-4' />
            Upload CSV
          </Button>
          <Button
            type='button'
            variant='outline'
            className='gap-2'
            disabled={
              !analysis ||
              saveCompanyConfigMutation.isPending ||
              (companyNames.length === 0 && fieldNames.length === 0)
            }
            onClick={saveCompanyConfig}
          >
            {saveCompanyConfigMutation.isPending ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <Save className='h-4 w-4' />
            )}
            Update Survey Data
          </Button>
        </div>

        {error && (
          <Alert variant='destructive'>
            <AlertCircle className='h-4 w-4' />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {successMessage && (
          <Alert>
            <AlertTitle>Ready</AlertTitle>
            <AlertDescription>{successMessage}</AlertDescription>
          </Alert>
        )}

        {analysis && (
          <div className='grid gap-4 md:grid-cols-3'>
            <div className='rounded-md border p-4'>
              <div className='text-sm text-muted-foreground'>Field alignment options</div>
              <div className='mt-3 flex flex-wrap gap-2'>
                {countEntries(analysis.fieldsOfBusiness).map(([field, count]) => (
                  <Badge key={field} variant='secondary'>
                    {field}: {count}
                  </Badge>
                ))}
              </div>
            </div>
            <div className='rounded-md border p-4'>
              <div className='text-sm text-muted-foreground'>Company sizes</div>
              <div className='mt-3 flex flex-wrap gap-2'>
                {countEntries(analysis.companySizes).map(([size, count]) => (
                  <Badge key={size} variant='secondary'>
                    {size}: {count}
                  </Badge>
                ))}
              </div>
            </div>
            <div className='rounded-md border p-4'>
              <div className='text-sm text-muted-foreground'>NDA projects</div>
              <div className='mt-2 text-2xl font-semibold'>{analysis.ndaRequiredCount}</div>
            </div>
          </div>
        )}

        {analysis && (
          <div className='rounded-md border'>
            <div className='border-b px-4 py-3 font-medium'>Constraint suggestions</div>
            <div className='divide-y'>
              {analysis.suggestions.map((suggestion) => (
                <div key={suggestion.id} className='p-4'>
                  <div className='flex flex-wrap items-center gap-2'>
                    <span className='font-medium'>{suggestion.title}</span>
                    <Badge variant='outline'>
                      {suggestion.lowerBound}-{suggestion.upperBound}
                    </Badge>
                    <Badge variant='secondary'>{suggestion.propertyName}</Badge>
                  </div>
                  <p className='mt-1 text-sm text-muted-foreground'>{suggestion.description}</p>
                  <p className='mt-2 text-xs text-muted-foreground'>
                    Applies to {suggestion.companyNames.length} hidden company projects.
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
