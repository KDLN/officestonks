import { supabase } from './supabase'
import { getEnvironmentConfig, logEnvironmentInfo } from '../config/environment'

// Sign up new user
export const signUp = async (email, password, userData = {}) => {
  try {
    const { data, error } = await supabase.auth.signUp({
      email,
      password,
      options: {
        data: {
          full_name: userData.full_name || userData.username,
          username: userData.username,
          ...userData
        },
        emailRedirectTo: getEnvironmentConfig().dashboardUrl
      }
    })

    if (error) throw error

    return {
      user: data.user,
      session: data.session,
      needsConfirmation: !data.session
    }
  } catch (error) {
    console.error('Sign up error:', error)
    throw error
  }
}

// Sign in existing user
export const signIn = async (email, password) => {
  try {
    const { data, error } = await supabase.auth.signInWithPassword({
      email,
      password
    })

    if (error) throw error

    return {
      user: data.user,
      session: data.session
    }
  } catch (error) {
    console.error('Sign in error:', error)
    throw error
  }
}

// Sign out
export const signOut = async () => {
  try {
    const { error } = await supabase.auth.signOut()
    if (error) throw error
  } catch (error) {
    console.error('Sign out error:', error)
    throw error
  }
}

// Get current user
export const getCurrentUser = async () => {
  try {
    const { data: { user }, error } = await supabase.auth.getUser()
    if (error) throw error
    return user
  } catch (error) {
    console.error('Get user error:', error)
    return null
  }
}

// Get current session
export const getCurrentSession = async () => {
  try {
    const { data: { session }, error } = await supabase.auth.getSession()
    if (error) throw error
    return session
  } catch (error) {
    console.error('Get session error:', error)
    return null
  }
}

// Check if user is authenticated
export const isAuthenticated = async () => {
  const session = await getCurrentSession()
  return !!session?.user
}

// Reset password
export const resetPassword = async (email) => {
  try {
    const { error } = await supabase.auth.resetPasswordForEmail(email, {
      redirectTo: `${getEnvironmentConfig().origin}/reset-password`
    })
    if (error) throw error
    return true
  } catch (error) {
    console.error('Reset password error:', error)
    throw error
  }
}

// Update password
export const updatePassword = async (newPassword) => {
  try {
    const { error } = await supabase.auth.updateUser({
      password: newPassword
    })
    if (error) throw error
    return true
  } catch (error) {
    console.error('Update password error:', error)
    throw error
  }
}

// Social auth providers
export const signInWithProvider = async (provider) => {
  try {
    const envConfig = logEnvironmentInfo()
    const redirectUrl = envConfig.dashboardUrl
    
    console.log(`${provider} OAuth redirect URL:`, redirectUrl)
    console.log('Environment:', envConfig.environment)
    
    const { data, error } = await supabase.auth.signInWithOAuth({
      provider,
      options: {
        redirectTo: redirectUrl
      }
    })
    if (error) throw error
    return data
  } catch (error) {
    console.error(`${provider} sign in error:`, error)
    console.error('Environment config:', getEnvironmentConfig())
    throw error
  }
}

// Auth state change listener
export const onAuthStateChange = (callback) => {
  return supabase.auth.onAuthStateChange(callback)
}

// Get user profile with additional data
export const getUserProfile = async (userId) => {
  try {
    const { data, error } = await supabase
      .from('profiles')
      .select('*')
      .eq('id', userId)
      .single()

    if (error && error.code !== 'PGRST116') throw error
    return data
  } catch (error) {
    console.error('Get user profile error:', error)
    return null
  }
}

// Update user profile
export const updateUserProfile = async (userId, updates) => {
  try {
    const { data, error } = await supabase
      .from('profiles')
      .upsert({ id: userId, ...updates })
      .select()
      .single()

    if (error) throw error
    return data
  } catch (error) {
    console.error('Update user profile error:', error)
    throw error
  }
}