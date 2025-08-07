# Repeater Usage Examples and Patterns

## Overview

This document provides comprehensive real-world examples of using repeater (`rpr`) for various continuous execution scenarios. Examples are organized by use case and demonstrate both basic and advanced usage patterns.

## Basic Usage Patterns

### 1. Health Monitoring

#### Simple Health Check
```bash
# Check service health every 30 seconds for 8 hours
rpr interval --every 30s --for 8h -- curl -f http://localhost:8080/health
```

#### Health Check with Logging
```bash
# Monitor with detailed logging and continue on failures
rpr interval --every 30s --for 8h --continue-on-error --output-file health.log -- \
    curl -f --max-time 10 http://api.example.com/health
```

#### Multi-Service Health Check
```bash
# Monitor multiple services in parallel
rpr interval --every 60s --for 24h --quiet -- bash -c '
    curl -f http://api.example.com/health && \
    curl -f http://db.example.com/health && \
    curl -f http://cache.example.com/health
'
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

This comprehensive set of examples demonstrates the versatility and power of repeater for various continuous execution scenarios, from simple monitoring to complex distributed workflows.