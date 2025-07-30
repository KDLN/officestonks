// Environment configuration helper
export const getEnvironmentConfig = () => {
  const hostname = window.location.hostname;
  const origin = window.location.origin;
  
  // Determine environment based on hostname
  const isBeta = hostname.includes('beta') || hostname.includes('dev');
  const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1';
  const isProduction = !isBeta && !isLocalhost;
  
  return {
    environment: isBeta ? 'beta' : isLocalhost ? 'development' : 'production',
    isBeta,
    isLocalhost, 
    isProduction,
    origin,
    dashboardUrl: `${origin}/dashboard`,
    // API URLs based on environment
    apiUrl: process.env.REACT_APP_API_URL || (
      isLocalhost ? 'http://localhost:8080' : 
      isBeta ? 'https://beta-api.officestonks.com' : // Will be configured later
      'https://api.officestonks.com' // Production API
    )
  };
};

// Console log environment info for debugging
export const logEnvironmentInfo = () => {
  const config = getEnvironmentConfig();
  console.log('Environment Config:', {
    ...config,
    timestamp: new Date().toISOString()
  });
  return config;
};