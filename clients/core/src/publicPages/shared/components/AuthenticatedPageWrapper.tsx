import { Footer } from './Footer'
import { Header } from './Header'
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  LoadingPage,
} from '@tumaet/prompt-ui-components'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import { useState } from 'react'

import DarkModeProvider from '@/contexts/DarkModeProvider'
import { useKeycloak } from '@core/keycloak/useKeycloak'

interface NonAuthenticatedPageWrapper {
  children: React.ReactNode
  withLoginButton?: boolean
}

// This page wrapper is meant for all pages that are only accessible to authenticated users, but which is not the management console.
export const AuthenticatedPageWrapper = ({
  children,
  withLoginButton = true,
}: NonAuthenticatedPageWrapper) => {
  const { keycloak } = useKeycloak()
  const { logout } = useAuthStore()
  const [isLogoutDialogOpen, setIsLogoutDialogOpen] = useState(false)

  const openLogoutDialog = () => setIsLogoutDialogOpen(true)
  const closeLogoutDialog = () => setIsLogoutDialogOpen(false)

  if (!keycloak) {
    return (
      <DarkModeProvider>
        <LoadingPage />
      </DarkModeProvider>
    )
  }
  return (
    <div className='min-h-screen flex flex-col'>
      <main className='flex-grow w-full px-4 sm:px-6 lg:px-8 py-12'>
        <div className='max-w-[1400px] mx-auto'>
          <Header withLoginButton={withLoginButton} onLogout={openLogoutDialog} />
          {children}
        </div>
      </main>
      <Footer />
      {/** Dialog to confirm logout */}
      <Dialog open={isLogoutDialogOpen} onOpenChange={setIsLogoutDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Logout</DialogTitle>
            <DialogDescription>
              Are you sure you want to log out? You progress will NOT be saved.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant='outline' onClick={closeLogoutDialog}>
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
