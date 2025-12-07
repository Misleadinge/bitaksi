# Testing Guide

This document explains the tests included in the TaxiHub project, how to run them, and how to analyze the results.

## Test Coverage Overview

The project maintains comprehensive test coverage across all layers:

**Current Coverage Status:**
-  **Driver Service Use Case**: 100% coverage
-  **Haversine Distance Calculation**: 100% coverage
-  **Driver Service Handler**: 91.1% coverage
-  **Driver Service Repository**: 83.7% coverage
-  **Gateway Handler**: 77.4% coverage
-  **Gateway Service Client**: 88.9% coverage

**Coverage Goals:**
- Critical business logic (use cases): Target 100%
- HTTP handlers: Target 90%+
- Repository/data access: Target 85%+
- Utilities (Haversine): Target 100%

## Included Tests

The project includes the following test suites:

### 1. Haversine Distance Calculation Tests
**Location:** `driver-service/pkg/haversine/haversine_test.go`

**What it tests:**
- Distance calculation between geographic coordinates using the Haversine formula
- Test cases include:
  - Istanbul to Ankara (~350 km)
  - Same point (should return 0 km)
  - Short distance (~1.4 km)

**Test function:** `TestDistance`

### 2. Driver Use Case Tests
**Location:** `driver-service/internal/usecase/driver_usecase_test.go`

**What it tests:**
- **Create Driver** (`TestDriverUseCase_CreateDriver`):
  - Valid driver creation
  - Validation errors (empty fields, invalid plate format, invalid taxi type, invalid location)
  
- **Update Driver** (`TestDriverUseCase_UpdateDriver`):
  - Partial updates (first name, location)
  - Driver not found error
  
- **List Drivers** (`TestDriverUseCase_ListDrivers`):
  - Pagination (page and pageSize)
  - Total count calculation
  - Empty results
  
- **Find Nearby Drivers** (`TestDriverUseCase_FindNearbyDrivers`):
  - Finding drivers within 6km radius
  - Filtering by taxi type
  - Invalid location coordinates
  - Empty results when no drivers nearby

**Test functions:**
- `TestDriverUseCase_CreateDriver`
- `TestDriverUseCase_UpdateDriver`
- `TestDriverUseCase_ListDrivers`
- `TestDriverUseCase_FindNearbyDrivers`

## Running Tests

### Run All Tests

From the project root:
```bash
make test
```

This runs all tests in both `driver-service` and `gateway` with verbose output.

### Run Tests Manually

```bash
# Run all tests in driver-service
cd driver-service
go test ./... -v

# Run all tests in gateway
cd gateway
go test ./... -v
```

### Run Specific Test Package

```bash
# Test haversine package
cd driver-service
go test ./pkg/haversine/... -v

# Test use case layer
cd driver-service
go test ./internal/usecase/... -v
```

### Run Specific Test Function

```bash
cd driver-service
go test ./internal/usecase/... -v -run TestDriverUseCase_CreateDriver
```

### Run Tests with Coverage

```bash
make test-coverage
```

This generates coverage reports for both services. Coverage files are saved as `coverage.out` in each service directory.

### Generate HTML Coverage Report

```bash
# For driver-service
cd driver-service
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Open coverage.html in your browser to see line-by-line coverage
```

### Run Tests with Race Detection

```bash
cd driver-service
go test ./... -race
```

This checks for race conditions in concurrent code.

## Analyzing Test Results

### Understanding Test Output

When running tests with `-v` flag, you'll see output like:

```
=== RUN   TestDistance
=== RUN   TestDistance/Istanbul_to_Ankara
--- PASS: TestDistance (0.00s)
    --- PASS: TestDistance/Istanbul_to_Ankara (0.00s)
    --- PASS: TestDistance/Same_point (0.00s)
    --- PASS: TestDistance/Short_distance (0.00s)
PASS
ok      github.com/bitaksi/driver-service/pkg/haversine    0.123s
```

**What to look for:**
- `PASS` - Test passed successfully
- `FAIL` - Test failed (check error message below)
- Test duration (e.g., `0.123s`) - How long the test took

### Failed Test Output

When a test fails, you'll see:

```
=== RUN   TestDriverUseCase_CreateDriver
=== RUN   TestDriverUseCase_CreateDriver/valid_driver
--- FAIL: TestDriverUseCase_CreateDriver (0.00s)
    --- FAIL: TestDriverUseCase_CreateDriver/valid_driver (0.00s)
        driver_usecase_test.go:123: CreateDriver() error = "plate must be in format: 2-3 digits, 1-3 letters, 1-4 digits", wantErr false
FAIL
```

**What to check:**
- The test name that failed
- The line number where the assertion failed (`driver_usecase_test.go:123`)
- The error message explaining what went wrong
- Expected vs actual values

### Coverage Analysis

After running tests with coverage:

```bash
# View coverage summary
cd driver-service
go tool cover -func=coverage.out
```

This shows coverage percentage for each function:

```
github.com/bitaksi/driver-service/internal/usecase/driver_usecase.go:73:	NewDriverUseCase	100.0%
github.com/bitaksi/driver-service/internal/usecase/driver_usecase.go:81:	CreateDriver		100.0%
github.com/bitaksi/driver-service/internal/usecase/driver_usecase.go:110:	UpdateDriver		100.0%
github.com/bitaksi/driver-service/internal/usecase/driver_usecase.go:176:	GetDriver		100.0%
github.com/bitaksi/driver-service/internal/usecase/driver_usecase.go:185:	ListDrivers		100.0%
github.com/bitaksi/driver-service/internal/usecase/driver_usecase.go:211:	FindNearbyDrivers	100.0%
total:										(statements)		100.0%
```

**What to look for:**
- Functions with low coverage (< 80%) may need more tests
- Functions with 0% coverage are not tested at all
- Overall coverage percentage at the bottom
- Per-function breakdown helps identify specific areas needing improvement

### Coverage by Package

To see coverage for specific packages:

```bash
# Driver Service - Use Case Layer
cd driver-service
go test ./internal/usecase/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -E "(usecase|total)"

# Driver Service - Handler Layer
go test ./internal/handler/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -E "(handler|total)"

# Driver Service - Repository Layer
go test ./internal/repository/mongodb/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -E "(repository|total)"

# Gateway - Handler Layer
cd ../gateway
go test ./internal/handler/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -E "(handler|total)"

# Gateway - Service Client
go test ./internal/service/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -E "(service|total)"
```

### Why cmd/ Shows 0% Coverage

When running coverage reports, you may see `cmd/` directories showing 0% coverage:

```
github.com/bitaksi/driver-service/cmd/driver-service    coverage: 0.0% of statements
github.com/bitaksi/gateway/cmd/gateway                  coverage: 0.0% of statements
```

**This is normal and expected!** The `cmd/` directories contain `main.go` files (application entry points) which are:

-  **Integration tested** - Tested via `test-services.sh` and Docker Compose
-  **Functionally verified** - Verified through end-to-end API testing
-  **Not unit tested** - By design, as they only wire up dependencies and start servers

**Best Practice:** Exclude `cmd/` from coverage requirements. Focus coverage efforts on:
- Business logic (`internal/usecase`) - Target 100%
- HTTP handlers (`internal/handler`) - Target 90%+
- Data access (`internal/repository`) - Target 85%+
- Utilities (`pkg/`) - Target 100%

**To exclude cmd/ from coverage reports:**

```bash
# Option 1: Test only internal and pkg packages
cd driver-service
go test ./internal/... ./pkg/... -coverprofile=coverage.out

# Option 2: Filter out cmd/ from coverage output
go tool cover -func=coverage.out | grep -v "cmd/"
```

### Understanding Coverage Percentages

**100% Coverage:**
- All code paths are tested
- Both success and error cases are covered
- Edge cases are handled

**80-99% Coverage:**
- Most code paths are tested
- May have some edge cases or error paths missing
- Generally acceptable for production code

**Below 80% Coverage:**
- Significant gaps in testing
- Should add more test cases
- Focus on error handling and edge cases

### Improving Coverage

To improve coverage for a specific function:

1. **Identify uncovered lines:**
   ```bash
   go tool cover -html=coverage.out -o coverage.html
   open coverage.html
   ```

2. **Look for red lines** in the HTML report - these are not covered

3. **Add test cases** for:
   - Error paths
   - Edge cases
   - Boundary conditions
   - Different input combinations

4. **Re-run coverage** to verify improvement:
   ```bash
   go test ./... -coverprofile=coverage.out
   go tool cover -func=coverage.out
   ```

### HTML Coverage Report

The HTML coverage report (`coverage.html`) shows:
- **Green lines**: Code covered by tests
- **Red lines**: Code not covered by tests
- **Gray lines**: Code not executable (comments, etc.)

Open the file in a browser to see which specific lines are covered.

## Integration Testing

For end-to-end testing of the entire system:

```bash
# Run the integration test script
./test-services.sh
```

This script verifies:
- Docker container health status
- Health check endpoints
- JWT authentication flow
- API endpoints (create, list, update, find nearby)
- Swagger documentation accessibility

**Expected output:**
- ✅ Green checkmarks for successful tests
- ❌ Red X marks for failed tests
- Error messages for any failures

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

### Tests Pass Locally but Fail in CI

- Check Go version compatibility
- Ensure all dependencies are in `go.mod`
- Verify environment variables are set correctly

### Coverage Report Not Generated

```bash
# Ensure you're in the correct directory
cd driver-service
# Run tests with coverage flag
go test ./... -coverprofile=coverage.out
# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```
