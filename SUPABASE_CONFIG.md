# Supabase Configuration for Beta Environment

## Issue: Discord OAuth redirects to main domain instead of beta domain

When using Discord login on `beta.officestonks.com`, users are redirected to the main `officestonks.com` domain instead of staying on the beta site.

## Root Cause

The Discord OAuth application in Supabase needs to have both domains configured as allowed redirect URLs.

## Required Supabase Configuration

### 1. Add Beta Domain to Discord OAuth Settings

In Supabase Dashboard:
1. Go to **Authentication > Providers**
2. Click on **Discord** provider
3. In **Redirect URLs**, add:
   - `https://beta.officestonks.com/dashboard` (for beta environment)
   - `https://officestonks.com/dashboard` (existing production)
   - `http://localhost:3000/dashboard` (for local development)

### 2. Environment Variables for Beta Deployment

The beta environment should use the same Supabase project but may need environment-specific variables:

```bash
# Same as production (shared Supabase project)
REACT_APP_SUPABASE_URL=your-supabase-url
REACT_APP_SUPABASE_ANON_KEY=your-anon-key

# Beta-specific API endpoints (when available)
REACT_APP_API_URL=https://beta-api.officestonks.com
```

### 3. Code Changes Made

- Added environment-aware redirect URL configuration
- Created `/config/environment.js` for environment detection
- Enhanced logging for debugging OAuth issues
- All auth functions now use environment-aware URLs

## Testing

After configuring Supabase:
1. Deploy to beta.officestonks.com
2. Try Discord login on beta site
3. Verify it redirects to `https://beta.officestonks.com/dashboard`
4. Check browser console for environment logging

## Alternative Solution

If using separate Supabase projects for beta/production is preferred:
1. Create separate Supabase project for beta
2. Configure Discord OAuth with beta domain only
3. Use different environment variables for beta deployment