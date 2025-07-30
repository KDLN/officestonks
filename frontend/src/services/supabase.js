// Supabase client - disabled for email auth
// Only initialize if explicitly configured with environment variables

let supabase = null

// Check if Supabase should be enabled
const supabaseUrl = process.env.REACT_APP_SUPABASE_URL
const supabaseAnonKey = process.env.REACT_APP_SUPABASE_ANON_KEY

if (supabaseUrl && supabaseAnonKey) {
  // Only import and initialize if env vars are present
  import('@supabase/supabase-js').then(({ createClient }) => {
    supabase = createClient(supabaseUrl, supabaseAnonKey, {
      auth: {
        autoRefreshToken: true,
        persistSession: true,
        detectSessionInUrl: true
      }
    })

    // Auth event listener for debugging
    supabase.auth.onAuthStateChange((event, session) => {
      console.log('Supabase auth event:', event, session?.user?.email)
    })
  }).catch(error => {
    console.error('Failed to initialize Supabase:', error)
  })
} else {
  console.log('Supabase disabled - using email authentication')
}

export { supabase }