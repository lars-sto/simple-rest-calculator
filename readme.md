# Simple Rest Calculator

This repository contains a simple REST calculator written in Go, showcasing idiomatic net/http handler design, 
explicit middleware composition (withLogging, withMemory, withRecover), and a lightweight in-memory history of 
recent calculations. The calculator operates on finite floating-point numbers within a defined numeric range
(|x| ≤ 1e15). Inputs or results outside this range are rejected with a client error.

## How to Run

```bash
go run ./cmd/calculator-service --addr :9090
```

## Persistence

By default, the recent calculation history is stored in memory only and is lost on restart.

You can enable persistence using a JSONL (JSON Lines) append-only log:

```bash
go run ./cmd/calculator-service --addr :9090 --persist --data-path ./data/recent.jsonl
```

### Compaction

On startup, the persistence file is compacted to keep only the last 20 entries.  
This prevents the JSONL log from growing indefinitely and keeps restarts fast.

## API Usage

### Notes
- All calculation endpoints expect JSON `{ "a": <int>, "b": <int> }`
- Division by zero returns `400 Bad Request`
- `/recent` returns the most recent calculations (default: 5, max: 20)


### Add numbers
```bash
curl -X POST http://localhost:9090/add \
  -H "Content-Type: application/json" \
  -d '{"a":1,"b":2}'
```

### Subtract numbers
```bash
curl -X POST http://localhost:9090/subtract \
-H "Content-Type: application/json" \
-d '{"a":5,"b":3}'
```

### Multiply numbers
```bash
curl -X POST http://localhost:9090/multiply \
  -H "Content-Type: application/json" \
  -d '{"a":4,"b":6}'
```

### Divide numbers
```bash
curl -X POST http://localhost:9090/divide \
  -H "Content-Type: application/json" \
  -d '{"a":10,"b":2}'
```

### Recent calculations
```bash
curl http://localhost:9090/recent?n=5
```

## Components
1. main
   - Wires all components together and starts the HTTP server
2. app
   - Holds shared dependencies and defines how handlers and middleware are composed
3. handler.go
   - Implements the actual HTTP endpoints: parse requests, call business logic, write responses
2. response_recorder
   - Wraps the ResponseWriter to capture status code and response body transparently
3. with_logging
   - Logs request and response metadata without affecting handler behavior
4. with_memory
   - Records successful operations and domain errors by observing inputs and outputs
5. with_recover
   - Catches panics and converts them into controlled HTTP 500 responses