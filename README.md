# Bitaksi TaxiHub

A production-grade microservice system for managing taxi drivers and querying nearby taxis, built with Go following Clean Architecture principles.

## Architecture

The system consists of two main services:

1. **API Gateway** - Single HTTP entry point that handles routing, authentication, rate limiting, and request forwarding
2. **Driver Service** - Microservice responsible for driver CRUD operations and nearby driver search

### Architecture Overview

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│         API Gateway                 │
│  - JWT Authentication              │
│  - Rate Limiting                   │
│  - Request Logging                 │
│  - Error Handling                  │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│      Driver Service                 │
│  - Driver CRUD Operations          │
│  - Nearby Driver Search            │
│  - Haversine Distance Calculation  │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│         MongoDB                     │
│  - Driver Data Storage             │
└─────────────────────────────────────┘
```

### Clean Architecture Layers

Both services follow Clean Architecture principles:

**Driver Service:**
- **Domain Layer** (`internal/domain`): Entities and business rules (no external dependencies)
- **Use Case Layer** (`internal/usecase`): Application logic and business workflows
- **Infrastructure Layer** (`internal/repository`, `internal/handler`): MongoDB implementation, HTTP handlers
- **Package Layer** (`pkg`): Shared utilities (Haversine calculation)

**Gateway:**
- **Handler Layer** (`internal/handler`): HTTP request handlers
- **Service Layer** (`internal/service`): Client for communicating with driver service
- **Middleware Layer** (`internal/middleware`): JWT auth, rate limiting, logging, error handling
- **Config Layer** (`internal/config`): Configuration management

## Features

### Core Features
- ✅ Driver CRUD operations (Create, Read, Update, List)
- ✅ Nearby driver search within 6km radius using Haversine formula
- ✅ Pagination support for driver listing
- ✅ Input validation and error handling
- ✅ Structured logging with correlation IDs
- ✅ MongoDB integration with proper connection handling

### Security & Performance
- ✅ JWT-based authentication (configurable)
- ✅ Rate limiting per IP address
- ✅ CORS support
- ✅ Request/response logging
- ✅ Global error handling
- ✅ Input validation

### Developer Experience
- ✅ Swagger/OpenAPI documentation
- ✅ Docker Compose setup
- ✅ Comprehensive unit tests
- ✅ Makefile with common commands
- ✅ Environment-based configuration

## Project Structure

```
bitaksi/
├── gateway/                    # API Gateway service
│   ├── cmd/
│   │   └── gateway/
│   │       └── main.go         # Gateway entry point
│   ├── internal/
│   │   ├── config/             # Configuration
│   │   ├── handler/            # HTTP handlers
│   │   ├── middleware/          # Middleware (JWT, rate limit, logging)
│   │   └── service/            # Driver service client
│   ├── docs/                   # Swagger documentation
│   ├── Dockerfile
│   └── go.mod
│
├── driver-service/              # Driver microservice
│   ├── cmd/
│   │   └── driver-service/
│   │       └── main.go         # Service entry point
│   ├── internal/
│   │   ├── domain/             # Domain entities
│   │   ├── usecase/            # Business logic
│   │   ├── repository/         # Data access interfaces
│   │   │   └── mongodb/        # MongoDB implementation
│   │   ├── handler/            # HTTP handlers
│   │   ├── middleware/          # Middleware
│   │   └── config/             # Configuration
│   ├── pkg/
│   │   └── haversine/          # Distance calculation utility
│   ├── docs/                   # Swagger documentation
│   ├── Dockerfile
│   └── go.mod
│
├── docker-compose.yml          # Docker Compose configuration
├── env.example                 # Environment variables example
├── .env                        # Your environment variables (create from env.example)
├── Makefile                    # Common commands
├── test-services.sh            # Integration test script
├── QUICKSTART.md              # Quick reference guide
└── README.md                   # This file
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- MongoDB (or use Docker Compose)
- Make (optional, for using Makefile commands)
- `swag` tool for Swagger generation (install with: `go install github.com/swaggo/swag/cmd/swag@latest`)

## Quick Start

### Using Docker Compose (Recommended)

1. **Clone and navigate to the project:**
   ```bash
   cd bitaksi
   ```

2. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

3. **Start all services:**
   ```bash
   docker-compose up --build
   ```

   This will start:
   - MongoDB on port 27017
   - Driver Service on port 8081
   - Gateway on port 8080

4. **Access services:**
   - Gateway: http://localhost:8080
   - Driver Service: http://localhost:8081
   - Gateway Swagger: http://localhost:8080/swagger/index.html
   - Driver Service Swagger: http://localhost:8081/swagger/index.html

### Running Locally (Without Docker)

1. **Start MongoDB:**
   ```bash
   # Using Docker
   docker run -d -p 27017:27017 --name mongodb mongo:7.0
   
   # Or use your local MongoDB instance
   ```

2. **Set environment variables:**
   ```bash
   export MONGODB_URI=mongodb://localhost:27017
   export MONGODB_DATABASE=taxihub
   export JWT_SECRET=your-secret-key
   export DRIVER_SERVICE_URL=http://localhost:8081
   ```

3. **Run driver service:**
   ```bash
   cd driver-service
   go mod download
   go run ./cmd/driver-service
   ```

4. **Run gateway (in another terminal):**
   ```bash
   cd gateway
   go mod download
   go run ./cmd/gateway
   ```

## API Endpoints

### Gateway Endpoints

All requests go through the gateway at `http://localhost:8080`

#### Authentication
- `POST /auth/login` - Login and get JWT token
  ```json
  {
    "username": "admin",
    "password": "password"
  }
  ```

#### Driver Management (Protected - requires JWT)
- `POST /drivers` - Create a new driver
- `PUT /drivers/:id` - Update a driver

#### Driver Queries (Public)
- `GET /drivers` - List drivers (with pagination)
- `GET /drivers/:id` - Get driver by ID
- `GET /drivers/nearby?lat=41.0082&lon=28.9784&taksiType=sari` - Find nearby drivers

### Example Requests

#### 1. Login to get JWT token:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**⚠️ Important:** When using the token in requests, you must include the "Bearer " prefix:
```bash
-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**For Swagger UI:** When entering your token in the "Authorize" dialog, you must manually add "Bearer " before the token (e.g., `Bearer eyJhbGci...`). Swagger UI 2.0 does not automatically add this prefix.

#### 2. Create a driver:
```bash
curl -X POST http://localhost:8080/drivers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "firstName": "Ahmet",
    "lastName": "Demir",
    "plate": "34ABC123",
    "taksiType": "sari",
    "carBrand": "Toyota",
    "carModel": "Corolla",
    "lat": 41.0431,
    "lon": 29.0099
  }'
```

#### 3. List drivers:
```bash
curl http://localhost:8080/drivers?page=1&pageSize=20
```

#### 4. Find nearby drivers:
```bash
curl "http://localhost:8080/drivers/nearby?lat=41.0082&lon=28.9784&taksiType=sari"
```

## Configuration

All configuration is done via environment variables. Docker Compose automatically reads `.env` file from the project root.

### Setting Up Environment Variables

1. **Copy the example file:**
   ```bash
   cp env.example .env
   ```

2. **Edit `.env` file** with your configuration values.

### ⚠️ Important: Docker Networking

When using Docker Compose, **always use service names** instead of `localhost`:

- ✅ **Correct**: `MONGODB_URI=mongodb://mongodb:27017` (uses Docker service name)
- ❌ **Wrong**: `MONGODB_URI=mongodb://localhost:27017` (won't work in containers)

- ✅ **Correct**: `DRIVER_SERVICE_URL=http://driver-service:8081` (for gateway)
- ❌ **Wrong**: `DRIVER_SERVICE_URL=http://localhost:8081` (won't work in containers)

### Key Configuration Variables

**MongoDB:**
- `MONGODB_URI` - MongoDB connection string (use `mongodb://mongodb:27017` for Docker)
- `MONGODB_DATABASE` - Database name (default: `taxihub`)

**JWT:**
- `JWT_SECRET` - Secret key for JWT signing (change in production!)
- `JWT_ENABLED` - Enable/disable JWT authentication (true/false)
- `JWT_EXPIRATION_HOURS` - Token expiration time in hours (default: 24)

**Rate Limiting:**
- `RATE_LIMIT_ENABLED` - Enable/disable rate limiting (default: true)
- `RATE_LIMIT_REQUESTS` - Number of requests allowed (default: 100)
- `RATE_LIMIT_WINDOW_SEC` - Time window in seconds (default: 60)

**Logging:**
- `LOG_LEVEL` - Log level (debug, info, warn, error, default: info)

**Service Ports:**
- `GATEWAY_PORT` - Gateway service port (default: 8080)
- `DRIVER_SERVICE_PORT` - Driver service port (default: 8081)

**Timeouts:**
- `READ_TIMEOUT_SEC` - HTTP read timeout in seconds (default: 30)
- `WRITE_TIMEOUT_SEC` - HTTP write timeout in seconds (default: 30)

## Testing

### Run all tests:
```bash
make test
```

### Run tests with coverage:
```bash
make test-coverage
```

### Run tests for a specific service:
```bash
cd driver-service && go test ./... -v
cd gateway && go test ./... -v
```

### Test Services (Integration Testing)

A test script is provided to verify all services are working:

```bash
# Run the test script
./test-services.sh
```

This script will:
- Check Docker container status
- Test health endpoints
- Test authentication
- Test API endpoints (create, list, find nearby)
- Verify Swagger documentation is accessible

## Development

### Common Commands

```bash
# Build both services
make build

# Run gateway locally
make run-gateway

# Run driver service locally
make run-driver-service

# Generate Swagger documentation
make swagger

# Run linter (requires golangci-lint)
make lint

# Tidy Go modules
make mod-tidy
```

### Generating Swagger Documentation

After making changes to API endpoints, regenerate Swagger docs:

```bash
# Install swag tool (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate for both services using Makefile (recommended)
make swagger

# Or generate individually
make swagger-driver
make swagger-gateway

# After generating, rebuild Docker containers
docker-compose build && docker-compose up -d
```

**Note:** Swagger docs are auto-generated from code annotations. Make sure to:
1. Add proper Swagger annotations to your handlers (see existing handlers for examples)
2. Run `make swagger` after adding/modifying endpoints
3. Rebuild containers to include updated documentation

## Design Patterns & Principles

### Clean Architecture
- **Separation of Concerns**: Domain logic is independent of infrastructure
- **Dependency Inversion**: High-level modules depend on abstractions (interfaces)
- **Testability**: Business logic can be tested without external dependencies

### Design Patterns Used
- **Repository Pattern**: Abstract data access layer
- **Dependency Injection**: Constructor injection for dependencies
- **Strategy Pattern**: Different distance calculation strategies (extensible)
- **Factory Pattern**: Service and handler creation

### SOLID Principles
- **Single Responsibility**: Each package/struct has one responsibility
- **Open/Closed**: Extensible through interfaces
- **Liskov Substitution**: Repository implementations are interchangeable
- **Interface Segregation**: Small, focused interfaces
- **Dependency Inversion**: Depend on abstractions, not concretions

## Error Handling

The system uses a consistent error response format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

### Error Codes
- `VALIDATION_ERROR` - Input validation failed
- `NOT_FOUND` - Resource not found
- `UNAUTHORIZED` - Authentication required or failed
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `INTERNAL_ERROR` - Server error

## Logging

Structured logging is implemented using `uber-go/zap`. Logs include:
- Request method, path, query parameters
- Response status code
- Request latency
- Client IP address
- Error details with context

Log level can be configured via `LOG_LEVEL` environment variable.

## Security Considerations

1. **JWT Authentication**: Configurable JWT-based auth for protected endpoints
2. **Rate Limiting**: Per-IP rate limiting to prevent abuse
3. **Input Validation**: All inputs are validated before processing
4. **Error Messages**: Internal errors are not exposed to clients
5. **CORS**: Configurable CORS headers
6. **Secrets Management**: All secrets come from environment variables

## Performance Considerations

1. **Connection Pooling**: MongoDB connection pooling
2. **Context Timeouts**: All database operations use context with timeouts
3. **Efficient Queries**: Indexed queries where applicable
4. **Rate Limiting**: Prevents service overload

## Troubleshooting

### Driver Service Can't Connect to MongoDB

**Symptom:** Driver service logs show `connection refused` or `failed to ping MongoDB`

**Solution:**
1. Check your `.env` file - ensure `MONGODB_URI` uses the service name:
   ```bash
   MONGODB_URI=mongodb://mongodb:27017  # ✅ Correct for Docker
   # NOT: mongodb://localhost:27017     # ❌ Wrong
   ```
2. Verify MongoDB container is running:
   ```bash
   docker-compose ps mongodb
   ```
3. Check MongoDB logs:
   ```bash
   docker-compose logs mongodb
   ```
4. Restart services:
   ```bash
   docker-compose down && docker-compose up -d
   ```

### Swagger Shows "No operations defined in spec!"

**Symptom:** Swagger UI is accessible but shows no endpoints

**Solution:**
1. Regenerate Swagger documentation:
   ```bash
   make swagger
   ```
2. Rebuild and restart containers:
   ```bash
   docker-compose build && docker-compose up -d
   ```

### Port Already in Use

**Symptom:** `Error: bind: address already in use`

**Solution:**
1. Change ports in `.env` file:
   ```bash
   GATEWAY_PORT=8082
   DRIVER_SERVICE_PORT=8083
   ```
2. Or stop conflicting services:
   ```bash
   # Find process using port
   lsof -i :8080
   # Kill the process
   kill -9 <PID>
   ```

### JWT Authentication Not Working

**Symptom:** Getting 401 Unauthorized even with valid token

**Solution:**
1. Ensure `JWT_SECRET` matches in both gateway and driver-service `.env`
2. Check if JWT is enabled:
   ```bash
   JWT_ENABLED=true
   ```
3. For testing, you can disable JWT:
   ```bash
   JWT_ENABLED=false
   ```

## Future Enhancements

Potential improvements for production:
- [ ] Redis for caching and rate limiting
- [ ] Database indexes for location queries
- [ ] Metrics and monitoring (Prometheus)
- [ ] Distributed tracing (Jaeger)
- [ ] Circuit breaker pattern for service communication
- [ ] Message queue for async operations
- [ ] Kubernetes deployment manifests
- [ ] CI/CD pipeline configuration

## License

Apache 2.0

## Contributing

This is a case study project. For production use, consider:
- Adding comprehensive integration tests
- Implementing proper user authentication (not mock)
- Adding database migrations
- Setting up monitoring and alerting
- Implementing API versioning
- Adding request/response compression
