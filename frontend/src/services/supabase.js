import { createClient } from '@supabase/supabase-js'

const supabaseUrl = process.env.REACT_APP_SUPABASE_URL
const supabaseAnonKey = process.env.REACT_APP_SUPABASE_ANON_KEY

if (!supabaseUrl || !supabaseAnonKey) {
  console.error('Missing Supabase environment variables. Please check your .env file.')
  console.log('Expected: REACT_APP_SUPABASE_URL and REACT_APP_SUPABASE_ANON_KEY')
}

// Create Supabase client with fallback values for testing environments
export const supabase = createClient(
  supabaseUrl || 'https://localhost.supabase.co', 
  supabaseAnonKey || 'test-anon-key', 
  {
    auth: {
      autoRefreshToken: true,
      persistSession: true,
      detectSessionInUrl: true
    }
  }
)

// Auth event listener for debugging
supabase.auth.onAuthStateChange((event, session) => {
  console.log('Supabase auth event:', event, session?.user?.email)
})