import React, { createContext, useContext, useEffect, useState } from 'react'
import { 
  signUp, 
  signIn, 
  signOut,
  getCurrentUser,
  getCurrentSession,
  onAuthStateChange,
  signInWithProvider,
  signInWithDiscordBeta
} from '../services/supabaseAuth'
import { syncAuthWithBackend } from '../services/authBridge'

const AuthContext = createContext({
  user: null,
  session: null,
  loading: true,
  signUp: () => {},
  signIn: () => {},
  signOut: () => {},
  signInWithProvider: () => {},
  signInWithDiscordBeta: () => {}
})

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null)
  const [session, setSession] = useState(null) 
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Get initial session
    const getInitialSession = async () => {
      try {
        const currentSession = await getCurrentSession()
        const currentUser = await getCurrentUser()
        
        setSession(currentSession)
        setUser(currentUser)
        
        // Beta redirect is now handled by AuthRedirect component
      } catch (error) {
        console.error('Error getting initial session:', error)
      } finally {
        setLoading(false)
      }
    }

    getInitialSession()

    // Listen for auth changes
    const { data: { subscription } } = onAuthStateChange(async (event, session) => {
      console.log('Auth state changed:', event, session?.user?.email)
      
      setSession(session)
      setUser(session?.user ?? null)
      
      // Sync with Office Stonks backend when user signs in
      if (event === 'SIGNED_IN' && session) {
        try {
          await syncAuthWithBackend()
          console.log('Successfully synced with Office Stonks backend')
          
          // Beta redirect is now handled by AuthRedirect component
        } catch (error) {
          console.error('Failed to sync with backend:', error)
        }
      }
      
      setLoading(false)
    })

    return () => {
      subscription?.unsubscribe()
    }
  }, [])

  const handleSignUp = async (email, password, userData = {}) => {
    try {
      const result = await signUp(email, password, userData)
      return result
    } catch (error) {
      throw error
    }
  }

  const handleSignIn = async (email, password) => {
    try {
      const result = await signIn(email, password)
      return result
    } catch (error) {
      throw error
    }
  }

  const handleSignOut = async () => {
    try {
      await signOut()
      setUser(null)
      setSession(null)
    } catch (error) {
      throw error
    }
  }

  const handleSignInWithProvider = async (provider) => {
    try {
      const result = await signInWithProvider(provider)
      return result
    } catch (error) {
      throw error
    }
  }

  const handleSignInWithDiscordBeta = async () => {
    try {
      const result = await signInWithDiscordBeta()
      return result
    } catch (error) {
      throw error
    }
  }

  const value = {
    user,
    session,
    loading,
    signUp: handleSignUp,
    signIn: handleSignIn,
    signOut: handleSignOut,
    signInWithProvider: handleSignInWithProvider,
    signInWithDiscordBeta: handleSignInWithDiscordBeta
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}