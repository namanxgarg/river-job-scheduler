FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the scheduler
RUN go build -o scheduler ./cmd/scheduler/main.go
# Build the worker
RUN go build -o worker ./cmd/worker/main.go

# By default, run scheduler if no command is specified
CMD ["./scheduler"]
