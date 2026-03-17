import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactNode } from 'react'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
})

interface ProvideProps {
  children: ReactNode
}

export const Provide = ({ children }: ProvideProps) => {
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}

export default Provide
