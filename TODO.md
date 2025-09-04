# NodeLocker TODO List

## 🚨 CRITICAL PRIORITY (Immediate Action Required)

### Security - Must Fix Before Production
- [ ] **URGENT: Move TLS certificates to secure storage**
  - Priority: Critical
  - Impact: Production security breach
  - Details: Current `/dev/shm` storage is temporary and insecure - certificates lost on reboot
  - Status: NOT IMPLEMENTED

- [ ] **URGENT: Complete password hashing migration**
  - Priority: Critical
  - Impact: Password security vulnerability
  - Details: Bcrypt support exists but SHA1 still used for new hashes - need to force bcrypt for all new registrations
  - Status: PARTIALLY IMPLEMENTED (upgrade mechanism exists, but not enforced)

## 🔴 HIGH PRIORITY (Next Sprint)

### Security Improvements
- [x] Implement rate limiting for API endpoints
  - Priority: High
  - Impact: Protection against brute force attacks
  - Details: Rate limiting middleware implemented with Redis backend
  - Status: COMPLETED

- [ ] Add secure headers
  - Priority: Medium
  - Impact: Improved security posture
  - Details: Implement security headers (HSTS, CSP, etc.)
  - Status: NOT IMPLEMENTED

### Code Quality
- [ ] Implement proper error handling
  - Priority: High
  - Impact: Reliability and debugging
  - Details: Remove `trunk-ignore` directives and handle all errors properly
  - Status: NOT IMPLEMENTED

- [ ] Implement configuration management
  - Priority: High
  - Impact: Deployment flexibility
  - Details: Add support for config files and environment variables (Redis URL, port, TLS settings)
  - Status: NOT IMPLEMENTED

- [ ] Refactor long functions
  - Priority: Medium
  - Impact: Maintainability
  - Details: Split `adminHandler` (150+ lines) and similar large functions
  - Status: NOT IMPLEMENTED

### Testing
- [ ] Add unit tests for core functionality
  - Priority: High
  - Impact: Code reliability
  - Details: Test password hashing, validation, Redis operations
  - Status: NOT IMPLEMENTED

- [ ] Implement integration tests
  - Priority: High
  - Impact: System reliability
  - Details: Add Redis integration tests with test database
  - Status: NOT IMPLEMENTED

### Documentation
- [ ] Add API documentation
  - Priority: High
  - Impact: Developer experience
  - Details: Implement Swagger/OpenAPI documentation for all endpoints
  - Status: NOT IMPLEMENTED

## 🟡 MEDIUM PRIORITY (Following Sprints)

### Code Quality
- [ ] Add structured logging
  - Priority: Medium
  - Impact: Observability
  - Details: Replace fmt.Println with proper logging framework (logrus, zap)
  - Status: NOT IMPLEMENTED

- [ ] Update HTTP methods
  - Priority: Medium
  - Impact: REST compliance
  - Details: Use proper POST/PUT/DELETE methods instead of GET for state changes
  - Status: NOT IMPLEMENTED

### Documentation
- [ ] Improve code documentation
  - Priority: Medium
  - Impact: Maintainability
  - Details: Add godoc comments for all exported functions
  - Status: PARTIALLY IMPLEMENTED (some functions documented)

- [ ] Create deployment guide
  - Priority: Medium
  - Impact: Operations
  - Details: Document production deployment steps, systemd config, Docker setup
  - Status: NOT IMPLEMENTED

- [ ] Add architecture documentation
  - Priority: Medium
  - Impact: System understanding
  - Details: Document system design, Redis data model, component interaction
  - Status: NOT IMPLEMENTED

### Testing
- [ ] Add API endpoint tests
  - Priority: Medium
  - Impact: API reliability
  - Details: Test all HTTP endpoints with various scenarios
  - Status: NOT IMPLEMENTED

- [ ] Create test environment
  - Priority: Medium
  - Impact: Testing reliability
  - Details: Setup isolated test environment with Docker Compose
  - Status: NOT IMPLEMENTED

### Operations
- [ ] Add health check endpoints
  - Priority: Medium
  - Impact: Operations
  - Details: Implement readiness and liveness probes (/health, /ready)
  - Status: NOT IMPLEMENTED

## 🟢 LOW PRIORITY (Future Releases)

### Monitoring & Performance
- [ ] Add metrics collection
  - Priority: Low
  - Impact: Monitoring
  - Details: Implement Prometheus metrics for requests, locks, errors
  - Status: NOT IMPLEMENTED

- [ ] Implement Redis connection pooling
  - Priority: Low
  - Impact: Performance
  - Details: Configure and optimize Redis connection pool settings
  - Status: NOT IMPLEMENTED

## ✅ COMPLETED ITEMS
- Rate limiting middleware implemented with Redis backend
- Bcrypt password support added (migration mechanism exists)
- Basic TLS support with self-signed certificates
- Core locking/unlocking functionality
- User registration and authentication
- Admin functions for environment management
- JSON and HTML status endpoints
- Redis data persistence with expiration
