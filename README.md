# NOTES

## Testing

In order to run user registration test to populate database with users follow next steps:

- Navigate to PocketArtisan/internal/modules/users/common/register
- run following command `  go clean -testcache && go test . -v -run TestBulkUserRegistration`

It is important to clean the cache before each retry since go caches previous test result



## Admin panel

- Idea is to first decide whether application for craftsman is approved or rejected
- Then use role set to set role in database