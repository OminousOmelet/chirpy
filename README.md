Chirpy is a simmple message posting server-side web API

ENDPOINTS
"admin/metrics"
  - GET, retreives data on number of visits to Chirpy
"admin/reset"
  - POST, 
"api/chirps
  - GET, retrieves all chirps, or just those from a specific user if an ID is passed by query
  - POST, posts a chirp from the request body
  - DELETE, delete a specific chirp by ID
"api/users"
  - POST, create a new user with info from request body
  - PUT, updates a user's credentials
"api/login"
  - POST, log in using a JWT token
"api/refresh"
  - POST, generate a backup refresh token for specified user
"api/revoke"
  - POST, revoke a refresh token for specified user
"api/polka/webhooks"
  - POST, upgrade user to Chirpy Red subscription
