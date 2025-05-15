# Distributed Job Scheduler

A high-performance, distributed job scheduling system built with Go, inspired by Celery and Sidekiq. This system supports asynchronous job processing with features like retry mechanisms, deduplication, and exponential backoff.

## Features

- Distributed job processing with multiple worker nodes
- Redis-based job queue for high throughput
- PostgreSQL for job metadata persistence
- Prometheus metrics for monitoring
- Kubernetes deployment support
- Docker containerization
- Job retry with exponential backoff
- Job deduplication
- High availability and scalability

## Architecture

The system consists of the following components:

1. **Scheduler**: Manages job scheduling and distribution
2. **Worker**: Processes jobs from the queue
3. **Queue**: Redis-based message queue
4. **Storage**: PostgreSQL for job metadata
5. **Monitoring**: Prometheus metrics and Grafana dashboards

## Prerequisites

- Go 1.21 or higher
- Redis 6.0 or higher
- PostgreSQL 13 or higher
- Docker
- Kubernetes cluster
- Helm 3

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/namangarg/job-scheduler.git
cd job-scheduler
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run the scheduler:
```bash
go run cmd/scheduler/main.go
```

5. Run a worker:
```bash
go run cmd/worker/main.go
```

## Docker Deployment

Build and run with Docker:

```bash
docker-compose up --build
```

## Kubernetes Deployment

Deploy to Kubernetes using Helm:

```bash
helm install job-scheduler ./helm/job-scheduler
```

## Monitoring

Access Prometheus metrics at `http://localhost:9090/metrics`

## License

MIT License 