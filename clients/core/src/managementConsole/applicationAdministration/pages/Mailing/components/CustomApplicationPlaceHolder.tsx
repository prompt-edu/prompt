import { AvailableMailPlaceholders, availablePlaceholders } from '@tumaet/prompt-ui-components'

const additionalMailPlacerholders = [
  {
    placeholder: '{{applicationEndDate}}',
    description:
      'The end date of the the application phase. (Not available for acceptance/rejection mails). ',
  },
  {
    placeholder: 'https://{{applicationURL}}',
    description:
      'The direct link to the application form. (Not available for acceptance/rejection mails).',
  },
]

export const applicationMailingPlaceholders = [
  ...availablePlaceholders,
  ...additionalMailPlacerholders,
]

export const CustomApplicationPlaceHolder = () => {
  return <AvailableMailPlaceholders customAdditionalPlaceholders={additionalMailPlacerholders} />
}
