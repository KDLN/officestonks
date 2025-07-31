import { useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { syncAuthWithBackend } from '../services/authBridge'
import { startTokenMonitoring, stopTokenMonitoring } from '../services/tokenManager'

const AuthSync = ({ children }) => {
  const { session, loading } = useAuth()

  useEffect(() => {
    const syncAuth = async () => {
      if (session?.access_token && !loading) {
        try {
          await syncAuthWithBackend()
          console.log('Auth synced with backend successfully')
          
          // Start token monitoring after successful sync
          startTokenMonitoring()
          console.log('🔐 Token monitoring started')
        } catch (error) {
          console.error('Failed to sync auth with backend:', error)
        }
      } else if (!session && !loading) {
        // Stop token monitoring when logged out
        stopTokenMonitoring()
        console.log('🔐 Token monitoring stopped')
      }
    }

    syncAuth()
  }, [session, loading])

  // Cleanup token monitoring on unmount
  useEffect(() => {
    return () => {
      stopTokenMonitoring()
    }
  }, [])

  return children
}

export default AuthSync