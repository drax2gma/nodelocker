# NodeLocker TODO List

## Security Improvements
- [ ] Replace SHA1 password hashing with bcrypt or Argon2
  - Priority: High
  - Impact: Critical for password security
  - Details: Current implementation uses static salts and SHA1

- [ ] Move TLS certificates to secure storage
  - Priority: High
  - Impact: Production security
  - Details: Current `/dev/shm` storage is temporary and insecure

- [ ] Implement rate limiting for API endpoints
  - Priority: High
  - Impact: Protection against brute force attacks
  - Details: Add rate limiting middleware using token bucket algorithm

- [ ] Add secure headers
  - Priority: Medium
  - Impact: Improved security posture
  - Details: Implement security headers (HSTS, CSP, etc.)

## Code Quality
- [ ] Implement proper error handling
  - Priority: High
  - Impact: Reliability and debugging
  - Details: Remove `trunk-ignore` directives and handle all errors

- [ ] Add structured logging
  - Priority: Medium
  - Impact: Observability
  - Details: Replace fmt.Println with proper logging framework

- [ ] Refactor long functions
  - Priority: Medium
  - Impact: Maintainability
  - Details: Split `adminHandler` and similar large functions

- [ ] Implement configuration management
  - Priority: High
  - Impact: Deployment flexibility
  - Details: Add support for config files and environment variables

- [ ] Update HTTP methods
  - Priority: Medium
  - Impact: REST compliance
  - Details: Use proper POST/PUT/DELETE methods instead of GET

## Documentation
- [ ] Add API documentation
  - Priority: High
  - Impact: Developer experience
  - Details: Implement Swagger/OpenAPI documentation

- [ ] Improve code documentation
  - Priority: Medium
  - Impact: Maintainability
  - Details: Add godoc comments for all exported functions

- [ ] Create deployment guide
  - Priority: Medium
  - Impact: Operations
  - Details: Document production deployment steps

- [ ] Add architecture documentation
  - Priority: Medium
  - Impact: System understanding
  - Details: Document system design and component interaction

## Testing
- [ ] Add unit tests for core functionality
  - Priority: High
  - Impact: Code reliability
  - Details: Test all core business logic functions

- [ ] Implement integration tests
  - Priority: High
  - Impact: System reliability
  - Details: Add Redis integration tests

- [ ] Add API endpoint tests
  - Priority: Medium
  - Impact: API reliability
  - Details: Test all HTTP endpoints

- [ ] Create test environment
  - Priority: Medium
  - Impact: Testing reliability
  - Details: Setup isolated test environment with Docker

## Future Enhancements
- [ ] Add metrics collection
  - Priority: Low
  - Impact: Monitoring
  - Details: Implement Prometheus metrics

- [ ] Add health check endpoints
  - Priority: Medium
  - Impact: Operations
  - Details: Implement readiness and liveness probes

- [ ] Implement Redis connection pooling
  - Priority: Low
  - Impact: Performance
  - Details: Configure and optimize Redis connection pool
