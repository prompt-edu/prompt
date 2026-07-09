import type { FeedbackItem } from '../../../interfaces/feedbackItem'

export const FeedbackSection = ({ feedbackItems }: { feedbackItems: FeedbackItem[] }) => {
  const positiveFeedback = feedbackItems.filter((item) => item.feedbackType === 'positive')
  const negativeFeedback = feedbackItems.filter((item) => item.feedbackType === 'negative')

  if (positiveFeedback.length === 0 && negativeFeedback.length === 0) return null

  return (
    <section className='mb-6 break-inside-avoid'>
      <h2 className='mb-2 border-b border-gray-200 pb-1 text-lg font-semibold'>Feedback</h2>
      {positiveFeedback.length > 0 && (
        <div className='mb-3'>
          <h3 className='text-sm font-medium'>What went well</h3>
          <ul className='ml-5 list-disc text-sm text-gray-700'>
            {positiveFeedback.map((item) => (
              <li key={item.id}>{item.feedbackText}</li>
            ))}
          </ul>
        </div>
      )}
      {negativeFeedback.length > 0 && (
        <div>
          <h3 className='text-sm font-medium'>Where to improve</h3>
          <ul className='ml-5 list-disc text-sm text-gray-700'>
            {negativeFeedback.map((item) => (
              <li key={item.id}>{item.feedbackText}</li>
            ))}
          </ul>
        </div>
      )}
    </section>
  )
}
