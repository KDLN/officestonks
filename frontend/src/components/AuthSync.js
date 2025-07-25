import { useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { syncAuthWithBackend } from '../services/authBridge'

const AuthSync = ({ children }) => {
  const { session, loading } = useAuth()

  useEffect(() => {
    const syncAuth = async () => {
      if (session?.access_token && !loading) {
        try {
          await syncAuthWithBackend()
          console.log('Auth synced with backend successfully')
        } catch (error) {
          console.error('Failed to sync auth with backend:', error)
        }
      }
    }

    syncAuth()
  }, [session, loading])

  return children
}

export default AuthSync