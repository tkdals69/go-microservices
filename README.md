# LiveOps Progression & Fairness Guard

A lightweight microservices suite for game LiveOps, providing event ingestion, fairness monitoring, progression tracking, and leaderboard management.

- **Author**: tkdals69
- **Cloud**: Azure (CLOUD=azure) 
- **Stack**: Go 1.22, chi, pgx, go-redis, prometheus
- **Architecture**: 4 microservices (gateway, fairness, progression, leaderboard)

## Services

### 1. Gateway Service (Port 8080)
- **POST /events**: Event ingestion with HMAC signature verification
- **Headers**: X-Signature, Idempotency-Key
- **Features**: Rate limiting, duplicate detection, message bus forwarding

### 2. Fairness Service (Port 8081) 
- **Monitoring**: Event flooding detection, anomalous score increases
- **Metrics**: dropped_events_total, anomaly_flags_total
- **Policy**: 429/403 responses for blocked players, audit logging

### 3. Progression Service (Port 8083)
- **Event Types**: progression, boss_kill, drop_claimed
- **Storage**: Redis (current season cache) + Postgres (append-only events)
- **POST /rewards/claim**: HMAC-signed reward receipt generation

### 4. Leaderboard Service (Port 8082)
- **Windows**: daily/weekly/seasonal, configurable top N (default 100)
- **Storage**: Redis ZSET pattern lb:{season}:{window}
- **GET /leaderboard**: Query with window/limit parameters
- **Snapshots**: Periodic Redis→Postgres archiving

## Quick Start

### Environment Setup (.env)
```env
CLOUD=azure
ENV=dev
PORT=8080
HMAC_SECRET=your-very-long-secret-key-at-least-32-characters-long
DB_URL=postgres://user:pass@host:5432/liveops?sslmode=require
REDIS_URL=rediss://:key@name.redis.cache.windows.net:6380/0
BUS_KIND=inmem
# See .env.example for full configuration
```

### Build & Run
```bash
# Install dependencies
go mod tidy

# Build all services
make build

# Run individual services
make run-gateway      # Port 8080
make run-progression  # Port 8083  
make run-leaderboard  # Port 8082
make run-fairness     # Port 8081

# Or run binaries directly
./bin/gateway
./bin/progression
./bin/leaderboard  
./bin/fairness
```

### Testing
```bash
# Run all tests
make test

# Test individual components
go test ./pkg/core/...
go test ./pkg/handlers/...
go test ./pkg/tests/...
```

## API Examples

### Event Ingestion
```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -H "X-Signature: your-hmac-signature" \
  -H "Idempotency-Key: unique-key-123" \
  -d @sample_events/progression.json
```

### Query Leaderboard
```bash
curl "http://localhost:8082/leaderboard/?window=weekly&limit=10"
```

### Claim Reward
```bash
curl -X POST http://localhost:8083/rewards/claim \
  -H "Content-Type: application/json" \
  -d '{"playerId":"player_123","item":"legendary_sword"}'
```

## Performance Targets
- **Latency**: p95 < 80ms @ 500 RPS (gateway)
- **Resources**: <70% CPU (2 vCPU), <300MB memory per service
- **Features**: Graceful shutdown (SIGTERM), 2s context timeout, exponential backoff retry

## Architecture

```
[Game Client] → [Gateway] → [Message Bus] → [Fairness/Progression/Leaderboard]
                    ↓
               [Audit Logs]
                    ↓
              [Metrics/Alerts]
```

## Message Bus Support
- **inmem**: Single-node in-memory channels (default)
- **sqs**: AWS Simple Queue Service  
- **servicebus**: Azure Service Bus

## Monitoring
- **/healthz**: Health check endpoints on all services
- **/metrics**: Prometheus metrics
- **Audit logs**: JSON structured logging to stdout

For detailed API documentation, see `api/openapi.yaml`.

## 샘플 이벤트 전송
```bash
BODY='{"type":"progression","playerId":"p1","ts":1730560000,"payload":{"deltaXp":10}}'
SIG="sha256=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "$HMAC_SECRET" -binary | xxd -p -c 256)"
curl -i -X POST "http://127.0.0.1:${PORT:-8080}/events" \
 -H "Content-Type: application/json" -H "X-Signature: $SIG" -H "Idempotency-Key: demo-1" -d "$BODY"
```

## Kubernetes 배포
- .env와 동일한 key=value로 Secret 생성
- 운영은 External Secrets(ESO) 권장

## 라이선스
MIT