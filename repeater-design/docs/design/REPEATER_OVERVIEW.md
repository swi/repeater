# Repeater (rpr) - Continuous Command Execution Tool

## Project Vision

**Repeater** is a command-line tool designed for continuous, scheduled, and rate-limited execution of commands. While traditional retry tools focus on making failed commands succeed, repeater focuses on running successful commands repeatedly with intelligent timing, rate limiting, and scheduling capabilities.

## Core Problem

Modern software operations require continuous execution patterns that existing tools don't handle well:

- **Monitoring**: Health checks every 30 seconds
- **Data Processing**: ETL jobs every hour  
- **API Polling**: Rate-limited API calls within quotas
- **Load Testing**: Sustained traffic generation
- **Maintenance**: Periodic cleanup tasks

Current solutions are inadequate:
- `watch` is too simple (fixed intervals only)
- `cron` is too rigid (time-based only, no rate limiting)
- Custom scripts are error-prone and lack standardization
- Retry tools stop after success (opposite of what we need)

## Solution: Repeater

Repeater provides intelligent continuous execution with:

1. **Flexible Scheduling**: Fixed intervals, rate limiting, adaptive timing
2. **Smart Stop Conditions**: Count-based, duration-based, or manual termination
3. **Rate Limiting**: Mathematical rate limiting to prevent quota violations
4. **Output Management**: Aggregation, filtering, and logging of repeated executions
5. **Enterprise Features**: Multi-instance coordination, metrics, observability

## Key Differentiators

| Tool | Use Case | Limitation |
|------|----------|------------|
| `watch` | Simple monitoring | Fixed intervals only |
| `cron` | Scheduled tasks | Time-based only, no rate limiting |
| `timeout` | Time-limited execution | Single execution only |
| Retry tools | Failure recovery | Stop after success |
| **repeater** | **Continuous execution** | **Designed for sustained operations** |

## Target Users

- **DevOps Engineers**: Monitoring, health checks, maintenance tasks
- **Data Engineers**: ETL pipelines, data processing workflows  
- **QA Engineers**: Load testing, sustained testing scenarios
- **API Developers**: Rate-limited API interactions, polling
- **System Administrators**: Periodic maintenance, log rotation

## Success Metrics

- **Adoption**: Used in production environments for critical operations
- **Reliability**: Zero missed executions due to tool failures
- **Efficiency**: Optimal resource utilization with rate limiting
- **Usability**: Intuitive CLI that reduces custom script development