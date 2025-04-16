# Golang Job Scheduler (Celery-like, Advanced)

A **production-style** job scheduling system in Go that uses:

- **Postgres** for storing job statuses  
- **Redis** as a broker to queue tasks  
- **Exponential backoff** (cenkalti/backoff) for task retries  
- **Multiple worker containers** for concurrency

Inspired by Celery in Python.

## Features

1. **Schedule Jobs** via POST `/jobs` with a JSON `{ "payload": "some data" }`.  
2. **Store** job metadata in Postgres (`jobs` table).  
3. **Workers** consume from Redis, run tasks with exponential backoff.  
4. **MaxRetries** logic to limit attempts if tasks keep failing.  
5. **Multiple workers** (worker1, worker2, etc.) for concurrency.  
6. **Docker Compose** setup to run `postgres`, `redis`, `scheduler`, and multiple `worker` containers.
