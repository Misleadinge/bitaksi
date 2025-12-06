# Testing Guide

This document explains how to run tests for the TaxiHub project.

## Quick Start

Run all tests:
```bash
make test
```

## Test Structure

The project includes the following tests:

### 1. Haversine Distance Calculation Tests
**Location:** `driver-service/pkg/haversine/haversine_test.go`

Tests the Haversine formula implementation for calculating distances between geographic coordinates.

**Run individually:**
```bash
cd driver-service
go test ./pkg/haversine/... -v
```

### 2. Driver Use Case Tests
**Location:** `driver-service/internal/usecase/driver_usecase_test.go`

Tests the business logic for driver operations:
- Create driver (with validation)
- Update driver
- List drivers
- Find nearby drivers

**Run individually:**
```bash
cd driver-service
go test ./internal/usecase/... -v
```

## Running Tests

### Run All Tests

```bash
# From project root
make test

# Or manually
cd driver-service && go test ./... -v
cd gateway && go test ./... -v
```

### Run Tests with Coverage

```bash
make test-coverage
```

This will generate coverage reports for both services.

### Run Specific Test Package

```bash
# Test haversine package
cd driver-service
go test ./pkg/haversine/... -v

# Test use case layer
cd driver-service
go test ./internal/usecase/... -v

# Test specific test function
cd driver-service
go test ./internal/usecase/... -v -run TestDriverUseCase_CreateDriver
```

### Run Tests with Verbose Output

```bash
go test ./... -v
```

The `-v` flag provides detailed output showing which tests pass or fail.

### Run Tests Multiple Times (Race Detection)

```bash
go test ./... -race -count=10
```

This runs tests 10 times and checks for race conditions.

## Test Coverage

Generate coverage report:

```bash
# For driver-service
cd driver-service
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Open coverage.html in browser to see coverage
```

## Integration Testing

For integration tests that verify the entire system:

```bash
# Run the integration test script
./test-services.sh
```

This script tests:
- Docker container health
- API endpoints
- Authentication
- Swagger documentation

## Writing New Tests

### Unit Test Example

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("MyFunction() = %v, expected %v", result, tt.expected)
            }
        })
    }
}
```

### Mock Example

The project uses interface-based mocking. See `driver_usecase_test.go` for an example of mocking the repository.

## Continuous Integration

For CI/CD pipelines, use:

```bash
# Run tests and fail on any failure
go test ./... -v -failfast

# Run tests with coverage threshold
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total | awk '{if ($3+0 < 80) exit 1}'
```

## Troubleshooting

### Tests Fail with "package not found"
```bash
# Ensure dependencies are downloaded
go mod download
go mod tidy
```

### Tests Fail with Import Errors
```bash
# Clean and rebuild
go clean -cache
go mod tidy
go test ./...
```

### Mock Not Working
- Ensure your mock implements all methods of the interface
- Check that the mock is properly initialized in test setup

## Best Practices

1. **Use Table-Driven Tests**: Makes tests easier to read and maintain
2. **Test Edge Cases**: Include boundary conditions and error cases
3. **Use Descriptive Test Names**: Test names should clearly describe what they test
4. **Keep Tests Fast**: Unit tests should run quickly (< 1 second)
5. **Isolate Tests**: Tests should not depend on each other
6. **Use Mocks**: Mock external dependencies (databases, APIs, etc.)

