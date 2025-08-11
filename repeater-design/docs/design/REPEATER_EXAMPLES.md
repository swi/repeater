# Repeater Usage Examples and Patterns

## Overview

This document provides comprehensive real-world examples of using repeater (`rpr`) for various continuous execution scenarios with Unix pipeline integration. Examples are organized by use case and demonstrate both basic and advanced usage patterns.

> **✅ Updated for v0.2.0 Unix Pipeline Ready** - All examples showcase the new pipeline-friendly behavior and output modes.

## Output Mode Examples

### Default Mode (Pipeline-Friendly)
Clean command output perfect for Unix pipelines:
```bash
# Count lines of output
rpr count --times 3 -- echo "test line" | wc -l
# Output: 3

# Extract specific data
rpr interval --every 2s --times 5 -- date +%H:%M:%S | tail -1
# Output: 14:23:45 (last timestamp)
```

### Quiet Mode
Suppress all command output, show only tool errors:
```bash
# Silent monitoring (only exit code matters)
rpr interval --every 30s --times 10 --quiet -- curl -f https://api.example.com
# No output unless there's an error

# Use in conditional scripts
if rpr count --times 3 --quiet -- ping -c 1 google.com; then
    echo "Network is up"
fi
```

### Verbose Mode
Full execution information plus command output:
```bash
# Detailed monitoring with full context
rpr interval --every 5s --times 3 --verbose -- curl -s https://api.example.com
# Shows execution info, command output, and final statistics
```

### Stats-Only Mode
Show only execution statistics:
```bash
# Performance monitoring
rpr count --times 100 --stats-only -- curl -w "%{time_total}\n" -o /dev/null -s https://api.example.com
# Shows only final statistics, no individual response times
```

## Unix Pipeline Integration Examples

### 1. Basic Pipeline Patterns

#### Count Successful Responses
```bash
# Monitor API and count successful responses
rpr interval --every 30s --times 20 -- curl -s -w "%{http_code}\n" -o /dev/null https://api.example.com | grep -c "200"
```

#### Extract and Process Data
```bash
# Get user data from API and extract specific fields
rpr count --times 5 -- curl -s https://api.github.com/user | jq -r '.login'
```

#### Monitor System Metrics
```bash
# Track disk usage over time
rpr duration --for 1h --every 5m -- df -h / | awk 'NR==2{print $5}' | tee disk-usage.log
```

#### Response Time Analysis
```bash
# Measure and analyze API response times
rpr count --times 50 -- curl -w "%{time_total}\n" -o /dev/null -s https://api.example.com | sort -n | tail -10
```

## Basic Usage Patterns

### 1. Health Monitoring

#### Simple Health Check with Pipeline
```bash
# Check service health and log status codes
rpr interval --every 30s --for 8h -- curl -s -w "%{http_code}\n" -o /dev/null http://localhost:8080/health | tee health.log
```

#### Health Check with Success Rate Calculation
```bash
# Monitor API health and calculate success rate
rpr interval --every 30s --times 100 -- curl -s -w "%{http_code}\n" -o /dev/null http://api.example.com/health | awk '$1==200{s++} END{print "Success rate:", s/NR*100"%"}'
```

#### Multi-Service Health Check with Aggregation
```bash
# Monitor multiple services and count healthy ones
rpr interval --every 60s --for 24h -- bash -c '
    services=0
    healthy=0
    
    if curl -f -s http://api.example.com/health > /dev/null; then
        healthy=$((healthy + 1))
    fi
    services=$((services + 1))
    
    if curl -f -s http://db.example.com/health > /dev/null; then
        healthy=$((healthy + 1))
    fi
    services=$((services + 1))
    
    if curl -f -s http://cache.example.com/health > /dev/null; then
        healthy=$((healthy + 1))
    fi
    services=$((services + 1))
    
    echo "$healthy/$services"
' | tee service-health.log
```

### 2. Data Processing

#### Periodic ETL Job
```bash
# Run ETL process every hour
rpr interval --every 1h --continue-on-error --timeout 30m -- \
    python /opt/etl/daily_sync.py
```

#### Batch Processing with Count Limit
```bash
# Process 100 batches with 5-minute intervals
rpr count --times 100 --every 5m --timeout 10m -- \
    ./process_batch.sh
```

#### Continuous Log Processing
```bash
# Process logs continuously for 12 hours
rpr duration --for 12h --every 30s --working-dir /var/log -- \
    ./log_processor.sh
```

### 3. API Polling

#### Rate-Limited API Polling
```bash
# Poll API respecting rate limits (future feature)
rpr rate-limit --limit 100/1h --continue-on-error -- \
    curl -H "Authorization: Bearer $TOKEN" https://api.example.com/data
```

#### Polling with Exponential Backoff on Failures
```bash
# Combine repeater with patience for resilient polling
rpr interval --every 60s --for 24h --continue-on-error -- \
    patience exponential --max-attempts 3 -- \
    curl -f https://api.example.com/updates
```

## DevOps and Monitoring Examples

### 1. Infrastructure Monitoring

#### System Resource Monitoring
```bash
# Monitor system resources every 5 minutes
rpr interval --every 5m --for 24h --output-file system-metrics.log -- bash -c '
    echo "=== $(date) ==="
    echo "CPU Usage:"
    top -l 1 | grep "CPU usage"
    echo "Memory Usage:"
    vm_stat | head -5
    echo "Disk Usage:"
    df -h /
    echo ""
'
```

#### Docker Container Health
```bash
# Monitor Docker containers every 30 seconds
rpr interval --every 30s --for 8h --continue-on-error -- bash -c '
    echo "=== Container Status $(date) ==="
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""
    
    # Check unhealthy containers
    unhealthy=$(docker ps --filter health=unhealthy --format "{{.Names}}")
    if [ -n "$unhealthy" ]; then
        echo "ALERT: Unhealthy containers: $unhealthy"
        exit 1
    fi
'
```

#### Kubernetes Pod Monitoring
```bash
# Monitor Kubernetes pods in production namespace
rpr interval --every 2m --for 12h --continue-on-error --output-file k8s-monitor.log -- bash -c '
    echo "=== Pod Status $(date) ==="
    kubectl get pods -n production --no-headers | grep -v Running | grep -v Completed
    
    # Check for pods in error states
    error_pods=$(kubectl get pods -n production --field-selector=status.phase!=Running,status.phase!=Succeeded --no-headers | wc -l)
    if [ "$error_pods" -gt 0 ]; then
        echo "ALERT: $error_pods pods in error state"
        kubectl get pods -n production --field-selector=status.phase!=Running,status.phase!=Succeeded
        exit 1
    fi
'
```

### 2. Database Maintenance

#### Database Health Checks
```bash
# Check database connectivity and performance
rpr interval --every 10m --for 24h --continue-on-error --timeout 30s -- bash -c '
    echo "=== Database Health $(date) ==="
    
    # Connection test
    psql -h db.example.com -U monitor -d production -c "SELECT 1;" > /dev/null
    if [ $? -eq 0 ]; then
        echo "✓ Database connection OK"
    else
        echo "✗ Database connection FAILED"
        exit 1
    fi
    
    # Performance check
    query_time=$(psql -h db.example.com -U monitor -d production -t -c "
        SELECT EXTRACT(EPOCH FROM (now() - query_start)) 
        FROM pg_stat_activity 
        WHERE state = '\''active'\'' 
        ORDER BY query_start 
        LIMIT 1;"
    )
    
    if (( $(echo "$query_time > 30" | bc -l) )); then
        echo "ALERT: Long running query detected: ${query_time}s"
    fi
'
```

#### Automated Database Cleanup
```bash
# Clean up old logs every 6 hours
rpr interval --every 6h --continue-on-error --timeout 1h -- bash -c '
    echo "=== Database Cleanup $(date) ==="
    
    # Clean old application logs (older than 30 days)
    deleted=$(psql -h db.example.com -U cleanup -d production -t -c "
        DELETE FROM application_logs 
        WHERE created_at < NOW() - INTERVAL '\''30 days'\'';
        SELECT ROW_COUNT();
    ")
    
    echo "Deleted $deleted old log entries"
    
    # Vacuum analyze after cleanup
    psql -h db.example.com -U cleanup -d production -c "VACUUM ANALYZE application_logs;"
    echo "Database maintenance completed"
'
```

### 3. Backup and Archival

#### Incremental Backups
```bash
# Perform incremental backups every 4 hours
rpr interval --every 4h --continue-on-error --timeout 2h --output-file backup.log -- bash -c '
    backup_dir="/backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    echo "=== Backup Started $(date) ==="
    echo "Backup directory: $backup_dir"
    
    # Database backup
    pg_dump -h db.example.com -U backup production > "$backup_dir/database.sql"
    
    # Application files backup
    rsync -av --exclude="*.log" /opt/app/ "$backup_dir/app/"
    
    # Compress backup
    tar -czf "$backup_dir.tar.gz" -C /backups "$(basename $backup_dir)"
    rm -rf "$backup_dir"
    
    # Upload to cloud storage
    aws s3 cp "$backup_dir.tar.gz" s3://backups/incremental/
    
    echo "=== Backup Completed $(date) ==="
    echo "Backup size: $(du -h $backup_dir.tar.gz | cut -f1)"
'
```

#### Log Rotation and Archival
```bash
# Archive and rotate logs daily
rpr interval --every 24h --immediate --continue-on-error -- bash -c '
    echo "=== Log Archival $(date) ==="
    
    # Archive application logs
    find /var/log/app -name "*.log" -mtime +1 -exec gzip {} \;
    find /var/log/app -name "*.log.gz" -mtime +30 -delete
    
    # Archive nginx logs
    find /var/log/nginx -name "*.log" -mtime +1 -exec gzip {} \;
    find /var/log/nginx -name "*.log.gz" -mtime +90 -delete
    
    # Send archived logs to long-term storage
    find /var/log -name "*.log.gz" -mtime +7 -exec aws s3 cp {} s3://log-archive/ \;
    find /var/log -name "*.log.gz" -mtime +7 -delete
    
    echo "Log archival completed"
'
```

## Load Testing and Performance Examples

### 1. API Load Testing

#### Sustained Load Test
```bash
# Generate sustained load for 1 hour
rpr duration --for 1h --every 100ms --continue-on-error --quiet -- \
    curl -w "%{http_code},%{time_total}\n" -o /dev/null -s \
    http://api.example.com/endpoint
```

#### Gradual Load Increase
```bash
# Start with low frequency, increase over time
rpr count --times 100 --every 5s --continue-on-error -- \
    curl -w "Phase1: %{http_code},%{time_total}\n" -o /dev/null -s \
    http://api.example.com/endpoint

# Then increase frequency
rpr count --times 200 --every 2s --continue-on-error -- \
    curl -w "Phase2: %{http_code},%{time_total}\n" -o /dev/null -s \
    http://api.example.com/endpoint

# Finally, peak load
rpr count --times 500 --every 500ms --continue-on-error -- \
    curl -w "Phase3: %{http_code},%{time_total}\n" -o /dev/null -s \
    http://api.example.com/endpoint
```

#### Parallel Load Testing
```bash
# Run multiple load generators in parallel
rpr count --times 1000 --every 1s --parallel 10 --continue-on-error -- bash -c '
    response=$(curl -w "%{http_code},%{time_total},%{size_download}" -o /dev/null -s \
        http://api.example.com/endpoint)
    echo "$(date +%s),$response"
'
```

### 2. Database Performance Testing

#### Connection Pool Testing
```bash
# Test database connection pool under load
rpr count --times 500 --every 200ms --parallel 20 --continue-on-error --timeout 10s -- bash -c '
    start_time=$(date +%s.%N)
    
    result=$(psql -h db.example.com -U test -d testdb -t -c "
        SELECT pg_sleep(0.1), NOW();
    " 2>&1)
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    if [[ $result == *"ERROR"* ]]; then
        echo "ERROR,$duration,$result"
        exit 1
    else
        echo "SUCCESS,$duration"
    fi
'
```

#### Query Performance Monitoring
```bash
# Monitor query performance during load test
rpr interval --every 10s --for 30m --output-file query-performance.log -- bash -c '
    echo "=== Query Performance $(date) ==="
    
    # Get slow queries
    psql -h db.example.com -U monitor -d production -c "
        SELECT query, calls, mean_time, total_time 
        FROM pg_stat_statements 
        WHERE mean_time > 1000 
        ORDER BY mean_time DESC 
        LIMIT 10;
    "
    
    # Get connection count
    connections=$(psql -h db.example.com -U monitor -d production -t -c "
        SELECT count(*) FROM pg_stat_activity WHERE state = '\''active'\'';
    ")
    echo "Active connections: $connections"
    
    echo ""
'
```

## Development and Testing Examples

### 1. Continuous Integration

#### Test Suite Execution
```bash
# Run test suite every 30 minutes during development
rpr interval --every 30m --for 8h --continue-on-error --output-file test-results.log -- bash -c '
    echo "=== Test Run $(date) ==="
    
    # Run unit tests
    npm test -- --reporter=json > test-results.json
    
    # Run integration tests
    npm run test:integration -- --reporter=json > integration-results.json
    
    # Check test results
    if [ $? -eq 0 ]; then
        echo "✓ All tests passed"
    else
        echo "✗ Tests failed"
        cat test-results.json | jq ".failures"
        exit 1
    fi
'
```

#### Code Quality Checks
```bash
# Run code quality checks every hour
rpr interval --every 1h --for 12h --continue-on-error -- bash -c '
    echo "=== Code Quality Check $(date) ==="
    
    # Linting
    npm run lint
    
    # Type checking
    npm run type-check
    
    # Security audit
    npm audit --audit-level=high
    
    # Test coverage
    npm run test:coverage
    coverage=$(cat coverage/coverage-summary.json | jq ".total.lines.pct")
    echo "Test coverage: $coverage%"
    
    if (( $(echo "$coverage < 80" | bc -l) )); then
        echo "WARNING: Test coverage below 80%"
    fi
'
```

### 2. Development Environment Monitoring

#### Hot Reload Monitoring
```bash
# Monitor file changes and trigger rebuilds
rpr interval --every 2s --continue-on-error --quiet -- bash -c '
    if [ -n "$(find src/ -newer .last-build 2>/dev/null)" ]; then
        echo "=== Changes detected, rebuilding $(date) ==="
        npm run build
        touch .last-build
        echo "Build completed"
    fi
'
```

#### Development Server Health
```bash
# Monitor development server and restart if needed
rpr interval --every 30s --continue-on-error -- bash -c '
    if ! curl -f http://localhost:3000/health > /dev/null 2>&1; then
        echo "=== Development server down, restarting $(date) ==="
        pkill -f "npm run dev"
        sleep 2
        npm run dev &
        sleep 5
        
        # Verify restart
        if curl -f http://localhost:3000/health > /dev/null 2>&1; then
            echo "✓ Development server restarted successfully"
        else
            echo "✗ Failed to restart development server"
            exit 1
        fi
    fi
'
```

## Advanced Integration Examples

### 1. Multi-Tool Coordination

#### Repeater + Patience for Resilient Monitoring
```bash
# Health check with retry logic and continuous monitoring
rpr interval --every 60s --for 24h --continue-on-error --daemon --resource-id health-monitor -- \
    patience exponential --max-attempts 3 --initial-delay 1s --max-delay 30s --daemon --resource-id health-retry -- \
    bash -c '
        # Comprehensive health check
        curl -f --max-time 10 http://api.example.com/health
        curl -f --max-time 5 http://api.example.com/ready
        curl -f --max-time 5 http://api.example.com/metrics
    '
```

#### Rate-Limited API Scraping
```bash
# Scrape API data with rate limiting coordination
rpr interval --every 30s --for 12h --daemon --resource-id api-scraper -- \
    patience linear --max-attempts 2 --daemon --resource-id api-retry -- \
    bash -c '
        # Fetch and process data
        data=$(curl -H "Authorization: Bearer $API_TOKEN" \
               https://api.example.com/data?limit=100)
        
        # Process and store data
        echo "$data" | jq ".items[]" | while read item; do
            # Store in database
            psql -h db.example.com -U app -d production -c \
                "INSERT INTO scraped_data (data, created_at) VALUES ('\''$item'\'', NOW());"
        done
        
        echo "Processed $(echo "$data" | jq ".items | length") items"
    '
```

### 2. Complex Workflows

#### Multi-Stage Data Pipeline
```bash
# Stage 1: Data extraction every 15 minutes
rpr interval --every 15m --for 24h --continue-on-error --output-file pipeline-extract.log -- bash -c '
    echo "=== Data Extraction $(date) ==="
    
    # Extract from multiple sources
    python /opt/pipeline/extract_api.py
    python /opt/pipeline/extract_files.py
    python /opt/pipeline/extract_database.py
    
    # Signal next stage
    touch /tmp/extraction_complete
'

# Stage 2: Data transformation (triggered by extraction)
rpr interval --every 1m --for 24h --continue-on-error --output-file pipeline-transform.log -- bash -c '
    if [ -f /tmp/extraction_complete ]; then
        echo "=== Data Transformation $(date) ==="
        
        python /opt/pipeline/transform.py
        
        # Clean up trigger and signal next stage
        rm /tmp/extraction_complete
        touch /tmp/transformation_complete
    fi
'

# Stage 3: Data loading (triggered by transformation)
rpr interval --every 1m --for 24h --continue-on-error --output-file pipeline-load.log -- bash -c '
    if [ -f /tmp/transformation_complete ]; then
        echo "=== Data Loading $(date) ==="
        
        python /opt/pipeline/load.py
        
        # Clean up trigger
        rm /tmp/transformation_complete
        
        echo "Pipeline completed successfully"
    fi
'
```

#### Distributed Processing Coordination
```bash
# Coordinator node
rpr interval --every 5m --for 8h --continue-on-error --daemon --resource-id job-coordinator -- bash -c '
    # Check for pending jobs
    pending_jobs=$(redis-cli llen job_queue)
    
    if [ "$pending_jobs" -gt 0 ]; then
        echo "=== Distributing $pending_jobs jobs $(date) ==="
        
        # Distribute jobs to workers
        for worker in worker1 worker2 worker3; do
            ssh $worker "rpr count --times 10 --every 1s --daemon --resource-id worker-$worker -- \
                bash -c '\''
                    job=\$(redis-cli lpop job_queue)
                    if [ -n \"\$job\" ]; then
                        echo \"Processing job: \$job\"
                        python /opt/worker/process_job.py \"\$job\"
                    fi
                '\''" &
        done
        
        wait
        echo "Job distribution completed"
    fi
'
```

## Monitoring and Alerting Examples

### 1. System Monitoring

#### Comprehensive System Monitor
```bash
# Monitor system health with alerting
rpr interval --every 5m --for 24h --continue-on-error --output-file system-monitor.log -- bash -c '
    echo "=== System Monitor $(date) ==="
    
    # CPU usage check
    cpu_usage=$(top -l 1 | grep "CPU usage" | awk "{print \$3}" | sed "s/%//")
    if (( $(echo "$cpu_usage > 80" | bc -l) )); then
        echo "ALERT: High CPU usage: $cpu_usage%"
        # Send alert
        curl -X POST https://hooks.slack.com/webhook \
             -d "{\"text\":\"High CPU usage: $cpu_usage%\"}"
    fi
    
    # Memory usage check
    memory_pressure=$(vm_stat | grep "Pages free" | awk "{print \$3}" | sed "s/\.//" )
    if [ "$memory_pressure" -lt 100000 ]; then
        echo "ALERT: Low memory: $memory_pressure pages free"
    fi
    
    # Disk usage check
    disk_usage=$(df / | tail -1 | awk "{print \$5}" | sed "s/%//")
    if [ "$disk_usage" -gt 85 ]; then
        echo "ALERT: High disk usage: $disk_usage%"
    fi
    
    # Network connectivity check
    if ! ping -c 1 8.8.8.8 > /dev/null 2>&1; then
        echo "ALERT: Network connectivity issue"
    fi
    
    echo "System check completed"
'
```

#### Application Performance Monitoring
```bash
# Monitor application performance metrics
rpr interval --every 2m --for 12h --continue-on-error --output-file app-performance.log -- bash -c '
    echo "=== Application Performance $(date) ==="
    
    # Response time check
    response_time=$(curl -w "%{time_total}" -o /dev/null -s http://localhost:8080/api/health)
    echo "Response time: ${response_time}s"
    
    if (( $(echo "$response_time > 2.0" | bc -l) )); then
        echo "ALERT: Slow response time: ${response_time}s"
    fi
    
    # Error rate check
    error_count=$(tail -100 /var/log/app/error.log | grep "$(date +%Y-%m-%d)" | wc -l)
    echo "Recent errors: $error_count"
    
    if [ "$error_count" -gt 10 ]; then
        echo "ALERT: High error rate: $error_count errors in last 100 log entries"
    fi
    
    # Memory usage check
    memory_usage=$(ps aux | grep "java.*myapp" | awk "{sum += \$6} END {print sum/1024}")
    echo "Memory usage: ${memory_usage}MB"
    
    if (( $(echo "$memory_usage > 2048" | bc -l) )); then
        echo "ALERT: High memory usage: ${memory_usage}MB"
    fi
'
```

### 2. Security Monitoring

#### Security Event Monitoring
```bash
# Monitor security events and suspicious activity
rpr interval --every 10m --for 24h --continue-on-error --output-file security-monitor.log -- bash -c '
    echo "=== Security Monitor $(date) ==="
    
    # Failed login attempts
    failed_logins=$(grep "Failed password" /var/log/auth.log | grep "$(date +%b\ %d)" | wc -l)
    echo "Failed login attempts today: $failed_logins"
    
    if [ "$failed_logins" -gt 50 ]; then
        echo "ALERT: High number of failed login attempts: $failed_logins"
        # Block suspicious IPs
        grep "Failed password" /var/log/auth.log | grep "$(date +%b\ %d)" | \
            awk "{print \$11}" | sort | uniq -c | sort -nr | head -5
    fi
    
    # Check for unusual network connections
    unusual_connections=$(netstat -an | grep ESTABLISHED | wc -l)
    echo "Active connections: $unusual_connections"
    
    if [ "$unusual_connections" -gt 100 ]; then
        echo "ALERT: Unusual number of network connections: $unusual_connections"
    fi
    
    # File integrity check
    if [ -f /etc/passwd.md5 ]; then
        current_md5=$(md5sum /etc/passwd | cut -d" " -f1)
        stored_md5=$(cat /etc/passwd.md5)
        
        if [ "$current_md5" != "$stored_md5" ]; then
            echo "ALERT: /etc/passwd has been modified!"
        fi
    else
        md5sum /etc/passwd | cut -d" " -f1 > /etc/passwd.md5
    fi
    
    echo "Security check completed"
'
```

## Configuration Examples

### 1. Configuration File Usage

#### Basic Configuration
```toml
# ~/.config/repeater/config.toml
[defaults]
continue_on_error = true
output_file = "/var/log/repeater.log"
timeout = "30s"
verbose = false

[interval]
jitter = "10%"
immediate = true

[count]
parallel = 3

[daemon]
socket_path = "/var/run/patience/daemon.sock"
enabled = true

[resources]
"api:external:api.example.com" = { limit = 100, window = "1h" }
"database:postgres:prod" = { limit = 50, window = "1m" }
```

#### Environment-Specific Configuration
```bash
# Development environment
export RPR_CONTINUE_ON_ERROR=true
export RPR_OUTPUT_FILE=/tmp/repeater-dev.log
export RPR_TIMEOUT=10s
export RPR_DAEMON_ENABLED=false

# Production environment
export RPR_CONTINUE_ON_ERROR=false
export RPR_OUTPUT_FILE=/var/log/repeater/production.log
export RPR_TIMEOUT=60s
export RPR_DAEMON_ENABLED=true
export RPR_RESOURCE_ID=production-monitor
```

### 2. Complex Scheduling Examples

#### Business Hours Monitoring
```bash
# Only run during business hours (9 AM - 5 PM, Mon-Fri)
rpr interval --every 5m --continue-on-error -- bash -c '
    hour=$(date +%H)
    day=$(date +%u)  # 1=Monday, 7=Sunday
    
    # Check if within business hours
    if [ "$day" -le 5 ] && [ "$hour" -ge 9 ] && [ "$hour" -lt 17 ]; then
        echo "=== Business Hours Monitor $(date) ==="
        curl -f http://api.example.com/business-metrics
    else
        echo "Outside business hours, skipping check"
    fi
'
```

#### Maintenance Window Awareness
```bash
# Skip monitoring during maintenance windows
rpr interval --every 1m --continue-on-error -- bash -c '
    # Check if maintenance window is active
    if [ -f /tmp/maintenance_mode ]; then
        echo "Maintenance mode active, skipping monitoring"
        exit 0
    fi
    
    # Check scheduled maintenance times
    hour=$(date +%H)
    if [ "$hour" -eq 2 ]; then  # 2 AM maintenance window
        echo "Scheduled maintenance window, skipping monitoring"
        exit 0
    fi
    
    # Normal monitoring
    echo "=== System Monitor $(date) ==="
    ./system_check.sh
'
```

## Best Practices and Patterns

### 1. Error Handling Patterns

#### Graceful Degradation
```bash
# Monitor with fallback mechanisms
rpr interval --every 30s --continue-on-error -- bash -c '
    # Primary check
    if curl -f http://primary.example.com/health > /dev/null 2>&1; then
        echo "✓ Primary service healthy"
        exit 0
    fi
    
    # Fallback check
    if curl -f http://secondary.example.com/health > /dev/null 2>&1; then
        echo "⚠ Primary down, secondary healthy"
        # Trigger failover if needed
        ./trigger_failover.sh
        exit 0
    fi
    
    # Both down
    echo "✗ Both primary and secondary services down"
    ./send_alert.sh "Critical: All services down"
    exit 1
'
```

#### Circuit Breaker Pattern
```bash
# Implement circuit breaker for external services
rpr interval --every 1m --continue-on-error -- bash -c '
    failure_file="/tmp/service_failures"
    
    # Check current failure count
    if [ -f "$failure_file" ]; then
        failures=$(cat "$failure_file")
    else
        failures=0
    fi
    
    # Circuit breaker logic
    if [ "$failures" -gt 5 ]; then
        echo "Circuit breaker open, skipping check"
        
        # Try to reset after 5 minutes
        if [ $(($(date +%s) - $(stat -c %Y "$failure_file"))) -gt 300 ]; then
            echo "Attempting to close circuit breaker"
            rm "$failure_file"
            failures=0
        else
            exit 0
        fi
    fi
    
    # Perform check
    if curl -f --max-time 10 http://external-api.example.com/health > /dev/null 2>&1; then
        echo "✓ External service healthy"
        # Reset failure count on success
        rm -f "$failure_file"
    else
        echo "✗ External service check failed"
        echo $((failures + 1)) > "$failure_file"
        exit 1
    fi
'
```

### 2. Resource Management

#### Memory-Conscious Processing
```bash
# Process large datasets in chunks
rpr count --times 100 --every 30s --continue-on-error --timeout 5m -- bash -c '
    echo "=== Processing batch $(date) ==="
    
    # Process in memory-efficient chunks
    python3 -c "
import gc
import sys

def process_chunk(chunk_id):
    # Process data chunk
    print(f\"Processing chunk {chunk_id}\")
    # ... processing logic ...
    
    # Explicit garbage collection
    gc.collect()
    
    # Memory usage check
    import psutil
    memory_percent = psutil.virtual_memory().percent
    if memory_percent > 80:
        print(f\"High memory usage: {memory_percent}%\")
        sys.exit(1)

# Process current chunk
chunk_id = $(date +%s)
process_chunk(chunk_id)
print(\"Chunk processing completed\")
"
'
```

#### Disk Space Management
```bash
# Monitor and manage disk space during processing
rpr interval --every 10m --continue-on-error -- bash -c '
    # Check available disk space
    available_space=$(df /var/log | tail -1 | awk "{print \$4}")
    
    # Convert to GB (assuming 1K blocks)
    available_gb=$((available_space / 1024 / 1024))
    
    echo "Available disk space: ${available_gb}GB"
    
    if [ "$available_gb" -lt 5 ]; then
        echo "Low disk space, cleaning up old files"
        
        # Clean up old log files
        find /var/log -name "*.log" -mtime +7 -delete
        find /var/log -name "*.log.gz" -mtime +30 -delete
        
        # Clean up temporary files
        find /tmp -type f -mtime +1 -delete
        
        echo "Cleanup completed"
    fi
    
    # Continue with normal processing
    ./process_logs.sh
'
```

## Load-Based Adaptive Scheduling Examples

### Overview

Load-based adaptive scheduling automatically adjusts execution intervals based on real-time system resource usage (CPU, memory, and load average). This feature is ideal for scenarios where system load varies significantly and you want to maintain optimal performance while avoiding system overload.

### 1. API Server Load Management

#### Adaptive API Health Monitoring
```bash
# Monitor API health with automatic load adjustment
rpr load-adaptive --base-interval 30s --target-cpu 70 --target-memory 80 --target-load 1.0 --for 24h -- bash -c '
    echo "=== API Health Check $(date) ==="
    
    # Comprehensive health check
    start_time=$(date +%s.%N)
    
    # Check main API endpoint
    api_response=$(curl -w "%{http_code},%{time_total}" -o /dev/null -s \
        --max-time 10 http://api.example.com/health)
    
    # Check database connectivity
    db_response=$(curl -w "%{http_code},%{time_total}" -o /dev/null -s \
        --max-time 5 http://api.example.com/db-health)
    
    # Check cache connectivity
    cache_response=$(curl -w "%{http_code},%{time_total}" -o /dev/null -s \
        --max-time 3 http://api.example.com/cache-health)
    
    end_time=$(date +%s.%N)
    total_time=$(echo "$end_time - $start_time" | bc)
    
    echo "API: $api_response, DB: $db_response, Cache: $cache_response"
    echo "Total check time: ${total_time}s"
    
    # Exit with error if any service is down
    if [[ $api_response != 200* ]] || [[ $db_response != 200* ]] || [[ $cache_response != 200* ]]; then
        echo "❌ One or more services unhealthy"
        exit 1
    fi
    
    echo "✅ All services healthy"
'
```

**Use Case**: During high traffic periods (Black Friday, product launches), the system automatically reduces monitoring frequency to avoid adding load. During low traffic, it increases frequency for better responsiveness.

#### Load-Aware Performance Testing
```bash
# Performance testing that adapts to system load
rpr load-adaptive --base-interval 1s --target-cpu 60 --target-memory 75 --target-load 0.8 --for 2h -- bash -c '
    echo "=== Load Test Iteration $(date) ==="
    
    # Generate controlled load based on current system capacity
    current_cpu=$(top -l 1 | grep "CPU usage" | awk "{print \$3}" | sed "s/%//")
    current_memory=$(vm_stat | grep "Pages active" | awk "{print \$3}" | sed "s/\.//" )
    
    echo "Current system state: CPU=${current_cpu}%, Memory pressure indicator=${current_memory}"
    
    # Adjust test intensity based on current load
    if (( $(echo "$current_cpu < 30" | bc -l) )); then
        # Low load - increase test intensity
        concurrent_requests=20
        request_timeout=5
    elif (( $(echo "$current_cpu < 60" | bc -l) )); then
        # Medium load - moderate intensity
        concurrent_requests=10
        request_timeout=10
    else
        # High load - reduce intensity
        concurrent_requests=5
        request_timeout=15
    fi
    
    echo "Test parameters: concurrent_requests=$concurrent_requests, timeout=${request_timeout}s"
    
    # Execute load test
    for i in $(seq 1 $concurrent_requests); do
        (
            response=$(curl -w "%{http_code},%{time_total},%{size_download}" \
                --max-time $request_timeout -o /dev/null -s \
                "http://api.example.com/load-test?iteration=$i")
            echo "Request $i: $response"
        ) &
    done
    
    wait
    echo "Load test iteration completed"
'
```

**Use Case**: Continuous performance testing that automatically scales back during peak hours to avoid impacting production traffic, and scales up during off-hours for comprehensive testing.

### 2. Database Operations and Maintenance

#### Adaptive Database Maintenance
```bash
# Database maintenance that adapts to system load
rpr load-adaptive --base-interval 15m --target-cpu 50 --target-memory 70 --target-load 1.2 --for 12h -- bash -c '
    echo "=== Database Maintenance $(date) ==="
    
    # Check current database load
    active_connections=$(psql -h db.example.com -U monitor -d production -t -c "
        SELECT count(*) FROM pg_stat_activity WHERE state = '\''active'\'';
    ")
    
    slow_queries=$(psql -h db.example.com -U monitor -d production -t -c "
        SELECT count(*) FROM pg_stat_activity 
        WHERE state = '\''active'\'' AND now() - query_start > interval '\''30 seconds'\'';
    ")
    
    echo "Database state: $active_connections active connections, $slow_queries slow queries"
    
    # Adjust maintenance operations based on database load
    if [ "$active_connections" -lt 10 ] && [ "$slow_queries" -eq 0 ]; then
        echo "Low database load - performing comprehensive maintenance"
        
        # Vacuum analyze on large tables
        psql -h db.example.com -U maintenance -d production -c "
            VACUUM ANALYZE user_sessions;
            VACUUM ANALYZE application_logs;
            VACUUM ANALYZE audit_trail;
        "
        
        # Update statistics
        psql -h db.example.com -U maintenance -d production -c "
            ANALYZE;
        "
        
        # Reindex if needed
        psql -h db.example.com -U maintenance -d production -c "
            REINDEX INDEX CONCURRENTLY idx_user_sessions_created_at;
        "
        
    elif [ "$active_connections" -lt 25 ]; then
        echo "Medium database load - performing light maintenance"
        
        # Only vacuum small tables
        psql -h db.example.com -U maintenance -d production -c "
            VACUUM user_preferences;
            VACUUM system_settings;
        "
        
    else
        echo "High database load - skipping maintenance"
        exit 0
    fi
    
    echo "Database maintenance completed"
'
```

**Use Case**: Database maintenance that automatically scales operations based on current database load. During peak business hours, it performs minimal maintenance. During off-hours, it performs comprehensive maintenance operations.

#### Load-Aware Data Processing Pipeline
```bash
# ETL pipeline that adapts processing intensity to system resources
rpr load-adaptive --base-interval 5m --target-cpu 65 --target-memory 80 --target-load 1.5 --for 8h -- bash -c '
    echo "=== ETL Pipeline $(date) ==="
    
    # Check available system resources
    available_memory=$(free | grep "Mem:" | awk "{print (\$7/\$2)*100}")
    cpu_idle=$(top -bn1 | grep "Cpu(s)" | awk "{print \$8}" | sed "s/%id,//")
    
    echo "System resources: ${available_memory}% memory available, ${cpu_idle}% CPU idle"
    
    # Determine batch size based on available resources
    if (( $(echo "$available_memory > 50 && $cpu_idle > 50" | bc -l) )); then
        batch_size=10000
        parallel_workers=8
        echo "High resources available - large batch processing"
    elif (( $(echo "$available_memory > 25 && $cpu_idle > 25" | bc -l) )); then
        batch_size=5000
        parallel_workers=4
        echo "Medium resources available - moderate batch processing"
    else
        batch_size=1000
        parallel_workers=2
        echo "Limited resources available - small batch processing"
    fi
    
    # Process data with adaptive batch size
    python3 -c "
import multiprocessing as mp
import psutil
import time
import sys

def process_batch(batch_id):
    print(f\"Processing batch {batch_id} with size $batch_size\")
    
    # Simulate data processing
    time.sleep(2)  # Simulated processing time
    
    # Monitor resource usage during processing
    cpu_percent = psutil.cpu_percent(interval=1)
    memory_percent = psutil.virtual_memory().percent
    
    if cpu_percent > 85 or memory_percent > 90:
        print(f\"Resource threshold exceeded: CPU={cpu_percent}%, Memory={memory_percent}%\")
        return False
    
    return True

# Process batches in parallel
with mp.Pool($parallel_workers) as pool:
    results = pool.map(process_batch, range(1, 6))  # Process 5 batches
    
    if all(results):
        print(\"All batches processed successfully\")
    else:
        print(\"Some batches failed due to resource constraints\")
        sys.exit(1)
"
    
    echo "ETL pipeline iteration completed"
'
```

**Use Case**: Data processing pipeline that automatically adjusts batch sizes and parallelism based on available system resources, ensuring optimal throughput without overwhelming the system.

### 3. Monitoring and Alerting Systems

#### Adaptive System Monitoring
```bash
# System monitoring with load-aware frequency adjustment
rpr load-adaptive --base-interval 2m --target-cpu 75 --target-memory 85 --target-load 2.0 --for 24h -- bash -c '
    echo "=== System Monitor $(date) ==="
    
    # Collect comprehensive system metrics
    cpu_usage=$(top -l 1 | grep "CPU usage" | awk "{print \$3}" | sed "s/%//")
    memory_usage=$(vm_stat | grep "Pages active" | awk "{print \$3}" | sed "s/\.//" )
    disk_usage=$(df / | tail -1 | awk "{print \$5}" | sed "s/%//")
    load_avg=$(uptime | awk -F"load averages:" "{print \$2}" | awk "{print \$1}")
    
    # Network connections
    tcp_connections=$(netstat -an | grep ESTABLISHED | wc -l)
    
    # Process count
    process_count=$(ps aux | wc -l)
    
    echo "System Metrics:"
    echo "  CPU Usage: ${cpu_usage}%"
    echo "  Memory Pressure: ${memory_usage}"
    echo "  Disk Usage: ${disk_usage}%"
    echo "  Load Average: ${load_avg}"
    echo "  TCP Connections: ${tcp_connections}"
    echo "  Process Count: ${process_count}"
    
    # Adaptive alerting based on system state
    alert_threshold_cpu=80
    alert_threshold_disk=90
    
    # Lower thresholds when system is already under stress
    if (( $(echo "$load_avg > 3.0" | bc -l) )); then
        alert_threshold_cpu=70
        alert_threshold_disk=85
        echo "System under stress - lowering alert thresholds"
    fi
    
    # Generate alerts
    alerts=()
    
    if (( $(echo "$cpu_usage > $alert_threshold_cpu" | bc -l) )); then
        alerts+=("High CPU usage: ${cpu_usage}%")
    fi
    
    if [ "$disk_usage" -gt "$alert_threshold_disk" ]; then
        alerts+=("High disk usage: ${disk_usage}%")
    fi
    
    if [ "$tcp_connections" -gt 1000 ]; then
        alerts+=("High connection count: ${tcp_connections}")
    fi
    
    # Send alerts if any
    if [ ${#alerts[@]} -gt 0 ]; then
        echo "🚨 ALERTS:"
        for alert in "${alerts[@]}"; do
            echo "  - $alert"
            # Send to monitoring system
            curl -X POST https://monitoring.example.com/alerts \
                 -H "Content-Type: application/json" \
                 -d "{\"message\": \"$alert\", \"timestamp\": \"$(date -Iseconds)\"}"
        done
    else
        echo "✅ All systems normal"
    fi
    
    # Log metrics to time series database
    curl -X POST https://metrics.example.com/api/v1/metrics \
         -H "Content-Type: application/json" \
         -d "{
             \"timestamp\": \"$(date -Iseconds)\",
             \"metrics\": {
                 \"cpu_usage\": $cpu_usage,
                 \"disk_usage\": $disk_usage,
                 \"load_average\": $load_avg,
                 \"tcp_connections\": $tcp_connections,
                 \"process_count\": $process_count
             }
         }"
'
```

**Use Case**: System monitoring that automatically reduces monitoring frequency during high load periods to avoid adding overhead, while increasing frequency during normal periods for better observability.

#### Load-Aware Log Analysis
```bash
# Log analysis that adapts processing intensity to system load
rpr load-adaptive --base-interval 10m --target-cpu 60 --target-memory 75 --for 12h -- bash -c '
    echo "=== Log Analysis $(date) ==="
    
    # Check log file sizes and system resources
    log_size=$(du -sm /var/log/application/*.log | awk "{sum += \$1} END {print sum}")
    available_memory_gb=$(free -g | grep "Mem:" | awk "{print \$7}")
    
    echo "Log files size: ${log_size}MB, Available memory: ${available_memory_gb}GB"
    
    # Adaptive processing strategy
    if [ "$available_memory_gb" -gt 4 ] && [ "$log_size" -lt 1000 ]; then
        # High memory, small logs - comprehensive analysis
        echo "Performing comprehensive log analysis"
        
        # Full text search for errors
        grep -r "ERROR\|FATAL\|CRITICAL" /var/log/application/ | \
            awk "{print \$1, \$2, \$3}" | sort | uniq -c | sort -nr > /tmp/error_summary.txt
        
        # Performance analysis
        grep -r "slow query\|timeout\|performance" /var/log/application/ | \
            wc -l > /tmp/performance_issues.txt
        
        # Security analysis
        grep -r "authentication failed\|unauthorized\|security" /var/log/application/ | \
            wc -l > /tmp/security_events.txt
        
        # Generate detailed report
        python3 -c "
import json
import datetime

# Read analysis results
with open('/tmp/error_summary.txt', 'r') as f:
    error_count = len(f.readlines())

with open('/tmp/performance_issues.txt', 'r') as f:
    perf_issues = int(f.read().strip())

with open('/tmp/security_events.txt', 'r') as f:
    security_events = int(f.read().strip())

# Generate report
report = {
    'timestamp': datetime.datetime.now().isoformat(),
    'analysis_type': 'comprehensive',
    'error_types': error_count,
    'performance_issues': perf_issues,
    'security_events': security_events,
    'log_size_mb': $log_size
}

print(json.dumps(report, indent=2))

# Send to monitoring system
import requests
requests.post('https://monitoring.example.com/log-analysis', json=report)
"
        
    elif [ "$available_memory_gb" -gt 2 ]; then
        # Medium memory - focused analysis
        echo "Performing focused log analysis"
        
        # Only check for critical errors
        critical_errors=$(grep -r "FATAL\|CRITICAL" /var/log/application/ | wc -l)
        recent_errors=$(grep -r "ERROR" /var/log/application/ | grep "$(date +%Y-%m-%d)" | wc -l)
        
        echo "Critical errors: $critical_errors, Recent errors: $recent_errors"
        
        if [ "$critical_errors" -gt 0 ] || [ "$recent_errors" -gt 100 ]; then
            echo "🚨 High error rate detected"
            # Send alert
            curl -X POST https://monitoring.example.com/alerts \
                 -d "{\"message\": \"High error rate: $critical_errors critical, $recent_errors recent\"}"
        fi
        
    else
        # Low memory - minimal analysis
        echo "Performing minimal log analysis"
        
        # Only check for critical system errors
        critical_count=$(grep -c "FATAL\|CRITICAL" /var/log/application/app.log 2>/dev/null || echo 0)
        
        if [ "$critical_count" -gt 0 ]; then
            echo "🚨 Critical errors found: $critical_count"
            # Send immediate alert
            curl -X POST https://monitoring.example.com/alerts \
                 -d "{\"message\": \"Critical errors detected: $critical_count\"}"
        fi
    fi
    
    echo "Log analysis completed"
'
```

**Use Case**: Log analysis system that adapts its processing depth based on available system resources. During high load, it performs minimal critical error checking. During low load, it performs comprehensive analysis including security and performance metrics.

### 4. Development and CI/CD Workflows

#### Adaptive Test Execution
```bash
# Test suite execution that adapts to system resources
rpr load-adaptive --base-interval 30m --target-cpu 70 --target-memory 80 --for 8h -- bash -c '
    echo "=== Adaptive Test Execution $(date) ==="
    
    # Check system resources and current load
    cpu_cores=$(nproc)
    available_memory_gb=$(free -g | grep "Mem:" | awk "{print \$7}")
    current_load=$(uptime | awk -F"load averages:" "{print \$2}" | awk "{print \$1}")
    
    echo "System: ${cpu_cores} cores, ${available_memory_gb}GB available memory, load: ${current_load}"
    
    # Determine test execution strategy
    if (( $(echo "$available_memory_gb > 8 && $current_load < 2.0" | bc -l) )); then
        # High resources - run full test suite with parallelization
        test_strategy="comprehensive"
        parallel_jobs=$((cpu_cores - 1))
        test_timeout="30m"
        
        echo "Running comprehensive test suite with $parallel_jobs parallel jobs"
        
        # Run all test categories
        npm run test:unit -- --parallel=$parallel_jobs --timeout=$test_timeout
        npm run test:integration -- --parallel=$((parallel_jobs / 2)) --timeout=$test_timeout
        npm run test:e2e -- --timeout=$test_timeout
        
        # Run performance tests
        npm run test:performance
        
        # Run security tests
        npm run test:security
        
    elif (( $(echo "$available_memory_gb > 4 && $current_load < 4.0" | bc -l) )); then
        # Medium resources - run core tests
        test_strategy="core"
        parallel_jobs=$((cpu_cores / 2))
        test_timeout="20m"
        
        echo "Running core test suite with $parallel_jobs parallel jobs"
        
        # Run essential tests only
        npm run test:unit -- --parallel=$parallel_jobs --timeout=$test_timeout
        npm run test:integration -- --timeout=$test_timeout
        
        # Skip performance and e2e tests
        echo "Skipping performance and e2e tests due to resource constraints"
        
    else
        # Low resources - run minimal critical tests
        test_strategy="minimal"
        parallel_jobs=1
        test_timeout="10m"
        
        echo "Running minimal test suite (single-threaded)"
        
        # Run only critical unit tests
        npm run test:unit -- --grep="critical" --timeout=$test_timeout
        
        echo "Skipping non-critical tests due to high system load"
    fi
    
    # Collect and report test results
    test_exit_code=$?
    
    # Generate test report
    python3 -c "
import json
import datetime
import subprocess

# Get test results
try:
    result = subprocess.run(['npm', 'run', 'test:report'], 
                          capture_output=True, text=True, timeout=60)
    test_results = json.loads(result.stdout) if result.stdout else {}
except:
    test_results = {'error': 'Failed to generate test report'}

# Create comprehensive report
report = {
    'timestamp': datetime.datetime.now().isoformat(),
    'strategy': '$test_strategy',
    'parallel_jobs': $parallel_jobs,
    'system_resources': {
        'cpu_cores': $cpu_cores,
        'available_memory_gb': $available_memory_gb,
        'current_load': float('$current_load')
    },
    'test_results': test_results,
    'exit_code': $test_exit_code
}

print(json.dumps(report, indent=2))

# Send to CI/CD system
import requests
try:
    requests.post('https://ci.example.com/test-reports', json=report, timeout=30)
    print('Test report sent to CI/CD system')
except:
    print('Failed to send test report')
"
    
    if [ $test_exit_code -eq 0 ]; then
        echo "✅ Tests passed with $test_strategy strategy"
    else
        echo "❌ Tests failed with $test_strategy strategy"
        exit 1
    fi
'
```

**Use Case**: CI/CD pipeline that adapts test execution based on available build server resources. During peak development hours with high server load, it runs minimal critical tests. During off-hours, it runs comprehensive test suites including performance and security tests.

#### Load-Aware Build Optimization
```bash
# Build system that adapts compilation settings to system resources
rpr load-adaptive --base-interval 15m --target-cpu 80 --target-memory 85 --for 4h -- bash -c '
    echo "=== Adaptive Build System $(date) ==="
    
    # Check for code changes
    if ! git diff --quiet HEAD~1; then
        echo "Code changes detected, starting adaptive build"
    else
        echo "No changes detected, skipping build"
        exit 0
    fi
    
    # Assess system resources
    cpu_cores=$(nproc)
    available_memory_gb=$(free -g | grep "Mem:" | awk "{print \$7}")
    disk_io_wait=$(iostat -c 1 2 | tail -1 | awk "{print \$4}")
    
    echo "Build environment: ${cpu_cores} cores, ${available_memory_gb}GB memory, ${disk_io_wait}% I/O wait"
    
    # Determine build configuration
    if (( $(echo "$available_memory_gb > 16 && $disk_io_wait < 10" | bc -l) )); then
        # High-performance build
        build_type="optimized"
        make_jobs=$((cpu_cores * 2))
        optimization_level="-O3"
        enable_lto="true"
        
        echo "High-performance build: $make_jobs parallel jobs, full optimization"
        
        # Configure build
        export MAKEFLAGS="-j$make_jobs"
        export CFLAGS="$optimization_level -flto"
        export CXXFLAGS="$optimization_level -flto"
        
        # Full build with all optimizations
        make clean
        make all
        make test
        
        # Generate optimized packages
        make package-optimized
        
    elif (( $(echo "$available_memory_gb > 8 && $disk_io_wait < 20" | bc -l) )); then
        # Standard build
        build_type="standard"
        make_jobs=$cpu_cores
        optimization_level="-O2"
        enable_lto="false"
        
        echo "Standard build: $make_jobs parallel jobs, standard optimization"
        
        # Configure build
        export MAKEFLAGS="-j$make_jobs"
        export CFLAGS="$optimization_level"
        export CXXFLAGS="$optimization_level"
        
        # Standard build
        make clean
        make all
        make test
        
    else
        # Minimal build
        build_type="minimal"
        make_jobs=1
        optimization_level="-O1"
        enable_lto="false"
        
        echo "Minimal build: single-threaded, basic optimization"
        
        # Configure build
        export MAKEFLAGS="-j1"
        export CFLAGS="$optimization_level"
        export CXXFLAGS="$optimization_level"
        
        # Incremental build only
        make
        
        # Skip tests if system is heavily loaded
        if (( $(echo "$disk_io_wait < 30" | bc -l) )); then
            make test-quick
        else
            echo "Skipping tests due to high I/O load"
        fi
    fi
    
    build_exit_code=$?
    
    # Generate build report
    build_time=$(date +%s)
    build_size=$(du -sh build/ | cut -f1)
    
    echo "Build completed: type=$build_type, time=${build_time}s, size=$build_size, exit_code=$build_exit_code"
    
    # Send build metrics
    curl -X POST https://ci.example.com/build-metrics \
         -H "Content-Type: application/json" \
         -d "{
             \"timestamp\": \"$(date -Iseconds)\",
             \"build_type\": \"$build_type\",
             \"make_jobs\": $make_jobs,
             \"optimization_level\": \"$optimization_level\",
             \"build_time\": $build_time,
             \"build_size\": \"$build_size\",
             \"exit_code\": $build_exit_code,
             \"system_resources\": {
                 \"cpu_cores\": $cpu_cores,
                 \"available_memory_gb\": $available_memory_gb,
                 \"disk_io_wait\": $disk_io_wait
             }
         }"
    
    if [ $build_exit_code -eq 0 ]; then
        echo "✅ Build successful"
    else
        echo "❌ Build failed"
        exit 1
    fi
'
```

**Use Case**: Build system that automatically adjusts compilation settings, parallelization, and optimization levels based on available system resources, ensuring optimal build times without overwhelming the build servers.

### 5. Best Practices for Load-Adaptive Scheduling

#### Resource Threshold Guidelines
```bash
# Conservative thresholds for production systems
rpr load-adaptive --base-interval 5m --target-cpu 60 --target-memory 70 --target-load 1.0

# Aggressive thresholds for development/testing
rpr load-adaptive --base-interval 1m --target-cpu 80 --target-memory 90 --target-load 2.0

# Balanced thresholds for mixed workloads
rpr load-adaptive --base-interval 2m --target-cpu 70 --target-memory 80 --target-load 1.5
```

#### Monitoring Load-Adaptive Behavior
```bash
# Monitor the adaptive scheduler's behavior
rpr load-adaptive --base-interval 1m --target-cpu 70 --show-metrics --for 1h -- bash -c '
    echo "=== Monitoring Adaptive Behavior $(date) ==="
    
    # Your actual workload here
    ./your-workload.sh
    
    # The --show-metrics flag will display:
    # - Current interval adjustments
    # - System resource utilization
    # - Load factor calculations
'
```

#### Combining with Other Scheduling Types
```bash
# Use load-adaptive for variable workloads
rpr load-adaptive --base-interval 5m --target-cpu 70 -- ./variable-workload.sh

# Use exponential backoff for external API calls
rpr backoff --initial 1s --max 60s -- curl external-api.com

# Use rate limiting for API quotas
rpr rate-limit --rate 100/1h -- curl rate-limited-api.com

# Use fixed intervals for predictable workloads
rpr interval --every 10m -- ./predictable-workload.sh
```

This comprehensive set of examples demonstrates the versatility and power of repeater for various continuous execution scenarios, from simple monitoring to complex distributed workflows.