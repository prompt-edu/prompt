import { useEffect, useState } from 'react'
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@tumaet/prompt-ui-components'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import { Footer } from './Footer'
import { Header } from './Header'

interface NonAuthenticatedPageWrapper {
  children: React.ReactNode
  withLoginButton?: boolean
}

export const NonAuthenticatedPageWrapper = ({
  children,
  withLoginButton = true,
}: NonAuthenticatedPageWrapper) => {
  const { logout } = useAuthStore()
  const [isLogoutDialogOpen, setIsLogoutDialogOpen] = useState(false)

  useEffect(() => {
    document.documentElement.classList.remove('dark')
  }, [])

  return (
    <div className='min-h-screen bg-white flex flex-col'>
      <main className='grow w-full px-4 sm:px-6 lg:px-8 py-12'>
        <div className='max-w-[1400px] mx-auto'>
          <Header withLoginButton={withLoginButton} onLogout={() => setIsLogoutDialogOpen(true)} />

          {children}
        </div>
      </main>
      <Footer />
      <Dialog open={isLogoutDialogOpen} onOpenChange={setIsLogoutDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Logout</DialogTitle>
            <DialogDescription>Are you sure you want to log out?</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant='outline' onClick={() => setIsLogoutDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant='destructive' onClick={() => logout()}>
              Logout
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
