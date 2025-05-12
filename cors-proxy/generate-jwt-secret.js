/**
 * This script generates a cryptographically secure random JWT secret.
 * Run with: node generate-jwt-secret.js
 */

const crypto = require('crypto');

// Generate a secure random secret of 64 bytes (512 bits) and encode as base64
const generateSecureSecret = () => {
  return crypto.randomBytes(64).toString('base64');
};

const jwtSecret = generateSecureSecret();

console.log('Generated JWT Secret:');
console.log(jwtSecret);
console.log('\nAdd this to your environment variables as:');
console.log(`JWT_SECRET=${jwtSecret}`);
console.log('\nIMPORTANT: Keep this secret secure and never commit it to source control!');