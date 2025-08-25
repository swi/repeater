# HTTPBin Real-World Testing Integration

This document describes the HTTPBin testing integration that provides real-world HTTP testing scenarios for the `rpr` (repeater) project. This testing framework enables validation of HTTP-aware retry strategies, error handling, and timing behaviors using live HTTP endpoints.

## Overview

The HTTPBin testing integration consists of:
- **Core testing utility** (`pkg/testing/httpbin.go`) - Network validation and common test scenarios
- **Integration tests** across 4 major components:
  - CLI integration tests (`cmd/rpr/main_httpbin_test.go`)
  - Runner integration tests (`pkg/runner/runner_httpbin_test.go`) 
  - Pattern matching tests (`pkg/patterns/patterns_httpbin_test.go`)
  - HTTP-aware performance benchmarks (`pkg/httpaware/httpaware_httpbin_bench_test.go`)
- **Smart conditional execution** - Tests automatically skip when offline or in CI environments

## Test Scenarios

The integration provides 7 real-world HTTP testing scenarios:

### 1. Service Unavailable (503)
- **Endpoint**: `https://httpbin.org/status/503`
- **Purpose**: Tests service temporary unavailability handling
- **Expected**: Exponential backoff retry behavior
- **Pattern**: `503|Service Unavailable`

### 2. Rate Limited (429) 
- **Endpoint**: `https://httpbin.org/status/429`
- **Purpose**: Tests rate limiting response handling
- **Expected**: Adaptive retry timing based on 429 responses
- **Pattern**: `429|Too Many Requests`

### 3. Server Error (502)
- **Endpoint**: `https://httpbin.org/status/502` 
- **Purpose**: Tests server error handling
- **Expected**: Exponential backoff with proper error detection
- **Pattern**: `502|Bad Gateway`

### 4. Success Response (200)
- **Endpoint**: `https://httpbin.org/status/200`
- **Purpose**: Tests successful execution completion
- **Expected**: Normal completion without retries
- **Pattern**: `200|OK`

### 5. JSON Response Parsing
- **Endpoint**: `https://httpbin.org/json`
- **Purpose**: Tests structured data parsing and pattern matching
- **Expected**: Successful JSON parsing and content extraction
- **Pattern**: `slideshow|origin`

### 6. Delayed Response
- **Endpoint**: `https://httpbin.org/delay/2`
- **Purpose**: Tests timing handling with delayed responses
- **Expected**: Proper timing measurement and HTTP-aware scheduling
- **Pattern**: `origin|args`

### 7. Headers Inspection
- **Endpoint**: `https://httpbin.org/headers`
- **Purpose**: Tests request header handling and inspection
- **Expected**: Proper header transmission and response parsing
- **Pattern**: `User-Agent|headers`

## Running HTTPBin Tests

### Prerequisites

1. **Network Connectivity**: Tests require internet access to reach `httpbin.org`
2. **Curl**: System must have `curl` available in PATH
3. **Go 1.19+**: Required for test execution

### Execution Commands

#### Run All HTTPBin Tests
```bash
# Run all HTTPBin integration tests
make test-integration

# Or manually with go test
go test -v --short=false ./... -run HTTPBin
```

#### Run Specific Test Suites

```bash
# CLI integration tests only
go test -v --short=false ./cmd/rpr/ -run HTTPBin

# Runner integration tests only  
go test -v --short=false ./pkg/runner/ -run HTTPBin

# Pattern matching tests only
go test -v --short=false ./pkg/patterns/ -run HTTPBin

# Performance benchmarks only
go test -v --short=false -bench=HTTPBin ./pkg/httpaware/
```

#### Run with Custom Configuration

```bash
# Run with extended timeout for slow networks
HTTBIN_TIMEOUT=30s go test -v ./... -run HTTPBin

# Skip network tests in CI/short mode
go test -v --short=true ./... -run HTTPBin  # Will skip automatically
```

### Test Behavior

#### Automatic Skipping
Tests automatically skip execution in these scenarios:
- **Short mode**: `go test -short` or `testing.Short() == true`
- **No connectivity**: When `httpbin.org` is unreachable
- **CI environments**: When network access is restricted
- **Timeout**: When HTTPBin responses exceed configured timeouts

#### Network Validation
Before running scenarios, tests perform:
1. Connectivity check to `https://httpbin.org/status/200`
2. Response time validation (< 3 second default)
3. Prerequisites validation (curl availability)

## Test Results and Expectations

### Success Criteria

#### CLI Integration Tests
- ✅ Command executes successfully with HTTPBin endpoints
- ✅ Verbose output contains strategy information
- ✅ HTTP-aware scheduling indicators present
- ✅ Proper retry timing for error responses

#### Runner Integration Tests  
- ✅ Executions complete with expected counts
- ✅ HTTP-aware timing adjustments applied
- ✅ Success/failure ratios match HTTP response codes
- ✅ Statistics accurately reflect execution results

#### Pattern Matching Tests
- ✅ JSON response parsing successful
- ✅ Pattern matching works with real HTTP responses
- ✅ Content extraction from structured data
- ✅ Header and metadata pattern recognition

#### Performance Benchmarks
- ✅ HTTP-aware scheduling performance measured
- ✅ Real-world timing scenarios benchmarked  
- ✅ Network latency impact quantified
- ✅ Memory usage with live HTTP operations

### Understanding Test Results

#### HTTP Status Code Behavior
**Important**: curl does not exit with non-zero status for HTTP error codes (4xx, 5xx) by default. This means:
- `httpbin.org/status/503` returns HTTP 503 but curl exits with code 0
- Tests measure HTTP response handling, not command failure
- Success/failure is determined by HTTP response analysis, not exit codes

#### Timing Expectations
- **Delayed responses**: Tests with `/delay/N` endpoints expect minimum N-second execution times
- **Retry scenarios**: Error status codes trigger retry behavior with exponential backoff
- **HTTP-aware**: Tests validate that retry timing considers HTTP response characteristics

## Troubleshooting

### Common Issues

#### 1. Tests Skipping Due to Network
```
HTTPBin not available - skipping real-world HTTP tests
```
**Solution**: Verify internet connectivity and `httpbin.org` accessibility:
```bash
curl -s https://httpbin.org/status/200
```

#### 2. Tests Skipping in Short Mode  
```
Skipping network tests in short mode
```
**Solution**: Run with network tests enabled:
```bash
go test -v --short=false ./... -run HTTPBin
```

#### 3. Curl Command Not Found
```
Cannot execute curl command
```
**Solution**: Install curl or verify it's in PATH:
```bash
which curl
curl --version
```

#### 4. Test Timeouts
```
context deadline exceeded
```
**Solution**: Increase timeout or check network performance:
```bash
# Increase timeout
HTTPBIN_TIMEOUT=30s go test ...

# Test network manually
time curl -s https://httpbin.org/json
```

### Advanced Configuration

#### Custom HTTPBin Instance
```go
// Use alternative HTTPBin instance
config := &httpbinTest.HTTPBinConfig{
    BaseURL: "https://httpbin.mycompany.com",
    ConnectTimeout: 5 * time.Second,
    RequestTimeout: 15 * time.Second, 
    MaxRetries: 3,
}
helper := httpbinTest.NewHTTPBinHelper(config)
```

#### Debug Mode
```bash
# Enable verbose output for debugging
go test -v -args -test.v=true ./... -run HTTPBin

# Show detailed timing information
HTTPBIN_DEBUG=1 go test -v ./... -run HTTPBin
```

## Integration with CI/CD

### GitHub Actions Example
```yaml
- name: Run HTTPBin Integration Tests
  run: |
    # Skip if no network access
    if curl -s --connect-timeout 5 https://httpbin.org/status/200; then
      go test -v --short=false ./... -run HTTPBin
    else
      echo "Skipping HTTPBin tests - no network access"
    fi
```

### Jenkins Pipeline
```groovy
stage('HTTPBin Integration') {
    steps {
        script {
            try {
                sh 'go test -v --short=false ./... -run HTTPBin'
            } catch (Exception e) {
                echo "HTTPBin tests skipped - network unavailable"
            }
        }
    }
}
```

## Architecture

### Core Components

#### HTTPBinHelper
- **Purpose**: Main testing utility with network validation
- **Features**: Connectivity checking, scenario generation, prerequisite validation
- **Configuration**: Timeouts, retries, base URL customization

#### HTTPBinEndpoints  
- **Purpose**: URL generation for common HTTPBin endpoints
- **Methods**: Status codes, delays, JSON, headers, user-agent
- **Flexibility**: Supports custom base URLs and parameter injection

#### TestScenario Structure
```go
type TestScenario struct {
    Name            string  // Unique scenario identifier
    Endpoint        string  // Full HTTPBin URL
    Method          string  // HTTP method (GET, POST, etc)
    ExpectedPattern string  // Regex pattern for response validation
    Description     string  // Human-readable description
}
```

### Design Principles

1. **Graceful Degradation**: Tests skip cleanly when network unavailable
2. **Real-World Validity**: Uses actual HTTP endpoints, not mocks
3. **CI/CD Friendly**: Designed for automated testing environments
4. **Performance Aware**: Includes benchmarking and timing validation
5. **Comprehensive Coverage**: Tests all major HTTP scenarios and error conditions

## Contributing

When adding new HTTPBin test scenarios:

1. **Add scenario to `CommonHTTPBinScenarios()`** in `pkg/testing/httpbin.go`
2. **Include test case** in relevant test files (`*_httpbin_test.go`)
3. **Update documentation** with new scenario details
4. **Test offline behavior** - ensure graceful skipping
5. **Validate CI compatibility** - test in automated environments

### Scenario Template
```go
{
    Name:            "your_scenario_name",
    Endpoint:        endpoints.YourEndpoint(),
    Method:          "GET",
    ExpectedPattern: "expected_response_pattern",
    Description:     "What this scenario tests and validates",
}
```

## Security Considerations

- **External Dependency**: Tests depend on `httpbin.org` external service
- **Network Exposure**: Tests make real HTTP requests to internet endpoints  
- **Data Transmission**: No sensitive data should be sent to HTTPBin endpoints
- **Rate Limiting**: HTTPBin may rate limit requests from CI systems
- **Availability**: External service availability may affect test reliability

## Performance Impact

- **Network I/O**: Tests perform real HTTP requests with network latency
- **Execution Time**: Complete suite takes 60-120 seconds depending on network
- **Resource Usage**: Minimal additional memory/CPU impact
- **Concurrency**: Tests run sequentially to avoid HTTPBin rate limiting

---

**Last Updated**: August 2024  
**Version**: 1.0.0  
**Author**: HTTPBin Integration Team