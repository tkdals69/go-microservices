# Go Microservices for LiveOps

[![Go Version](https://img.shields.io/badge/Go-1.24.1-blue.svg)](https://golang.org)
[![Azure](https://img.shields.io/badge/Cloud-Azure-blue.svg)](https://azure.microsoft.com)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Production-ready microservices suite for game LiveOps and real-time event processing.**

A scalable, cloud-native microservices architecture built with Go, designed for game LiveOps scenarios including event ingestion, fairness monitoring, progression tracking, and leaderboard management.

## 🏗️ Architecture

### Services Overview

| Service | Port | Description | Key Features |
|---------|------|-------------|--------------|
| **Gateway** | 8080 | Event ingestion & routing | HMAC auth, rate limiting, idempotency |
| **Fairness** | 8081 | Anti-cheat & anomaly detection | Event flooding detection, score validation |
| **Leaderboard** | 8082 | Ranking & competitions | Real-time rankings, seasonal windows |
| **Progression** | 8083 | Player progress tracking | Reward claims, achievement unlocks |
| **Web UI** | 3000 | Dashboard & monitoring | Real-time metrics, admin interface |

### Technology Stack

- **Language**: Go 1.24.1
- **Web Framework**: Chi v5 (lightweight, fast routing)
- **Database**: PostgreSQL (primary), Redis (cache/sessions)
- **Message Bus**: Azure Service Bus / In-memory (configurable)
- **Monitoring**: Prometheus metrics, structured logging
- **Cloud**: Azure-native (Blob Storage, Service Bus, Cache for Redis)
- **Containerization**: Docker + Docker Compose

## 🚀 Quick Start

### Prerequisites

- Go 1.24.1+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+
- Azure Account (for production deployment)

### 1. Clone & Setup

```bash
git clone https://github.com/tkdals69/go-microservices.git
cd go-microservices

# Copy environment template
cp .env.example .env
```

### 2. Configure Environment

Edit `.env` file with your settings:

```env
# Core Configuration
CLOUD=azure
ENV=development
HMAC_SECRET=your-very-long-secret-key-at-least-32-characters-long

# Database
DB_URL=postgres://user:pass@localhost:5432/liveops?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0

# Azure Configuration (for production)
AZURE_STORAGE_CONNECTION_STRING=DefaultEndpointsProtocol=https;AccountName=...
AZURE_SERVICE_BUS_CONNECTION_STRING=Endpoint=sb://...

# Message Bus
BUS_KIND=inmem  # or azure_servicebus for production

# Observability
LOG_LEVEL=info
METRICS_ENABLED=true
```

### 3. Build & Run

#### Option A: Using Make (Recommended)

```bash
# Build all services
make build

# Run individual services
make run-gateway
make run-fairness
make run-leaderboard
make run-progression
make run-web

# Or run all with Docker Compose
make docker-up
```

#### Option B: Manual Build

```bash
# Install dependencies
go mod tidy

# Build all services
go build -o bin/gateway ./cmd/gateway
go build -o bin/fairness ./cmd/fairness
go build -o bin/leaderboard ./cmd/leaderboard
go build -o bin/progression ./cmd/progression
go build -o bin/web ./web

# Run services
./bin/gateway &
./bin/fairness &
./bin/leaderboard &
./bin/progression &
./bin/web &
```

#### Option C: Docker Compose

```bash
# Start all services + dependencies
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### 4. Verify Installation

```bash
# Check service health
curl http://localhost:8080/health  # Gateway
curl http://localhost:8081/health  # Fairness
curl http://localhost:8082/health  # Leaderboard
curl http://localhost:8083/health  # Progression

# Open Web UI
open http://localhost:3000
```

## 📊 API Documentation

### Gateway Service (Port 8080)

#### Event Ingestion
```http
POST /events
Content-Type: application/json
X-Signature: sha256=<hmac>
Idempotency-Key: <uuid>

{
  "player_id": "player123",
  "event_type": "progression",
  "data": {
    "level": 5,
    "score": 1000
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Fairness Service (Port 8081)

#### Metrics Endpoint
```http
GET /metrics
# Returns Prometheus metrics
```

### Leaderboard Service (Port 8082)

#### Get Rankings
```http
GET /leaderboard?window=daily&limit=10

Response:
{
  "window": "daily",
  "season": "2024-Q1",
  "rankings": [
    {"rank": 1, "player_id": "player456", "score": 9999},
    {"rank": 2, "player_id": "player123", "score": 8888}
  ]
}
```

### Progression Service (Port 8083)

#### Claim Rewards
```http
POST /rewards/claim
Content-Type: application/json
X-Signature: sha256=<hmac>

{
  "player_id": "player123",
  "reward_type": "boss_kill",
  "reward_data": {"boss_id": "dragon", "loot": ["sword", "gold"]}
}
```

## 🛠️ Development

### Project Structure

```
├── cmd/                    # Service entry points
│   ├── gateway/           # Gateway service
│   ├── fairness/          # Fairness service
│   ├── leaderboard/       # Leaderboard service
│   └── progression/       # Progression service
├── pkg/                   # Shared packages
│   ├── adapters/          # External service adapters
│   │   ├── cache/         # Redis adapter
│   │   └── cloud/         # Azure adapters
│   ├── config/            # Configuration management
│   ├── core/              # Business logic
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── observability/     # Logging & metrics
│   └── tests/             # Test utilities
├── web/                   # Web UI
│   ├── static/            # CSS, JS assets
│   └── templates/         # HTML templates
├── api/                   # OpenAPI specifications
├── sample_events/         # Test event data
├── .github/              # CI/CD workflows
└── docs/                 # Additional documentation
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Benchmark tests
make benchmark
```

### Code Quality

```bash
# Format code
make format

# Lint code
make lint

# Security scan
make security-scan

# Generate documentation
make docs
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `CLOUD` | Cloud provider (azure/local) | `local` | ✅ |
| `ENV` | Environment (development/staging/production) | `development` | ✅ |
| `HMAC_SECRET` | HMAC signing secret (32+ chars) | - | ✅ |
| `DB_URL` | PostgreSQL connection string | - | ✅ |
| `REDIS_URL` | Redis connection string | - | ✅ |
| `BUS_KIND` | Message bus type (inmem/azure_servicebus) | `inmem` | ❌ |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` | ❌ |
| `METRICS_ENABLED` | Enable Prometheus metrics | `true` | ❌ |

### Service-Specific Ports

- Gateway: `8080`
- Fairness: `8081`
- Leaderboard: `8082`
- Progression: `8083`
- Web UI: `3000`
- Prometheus: `9090` (if enabled)

## 🚢 Deployment

### Local Development

Use Docker Compose for local development with all dependencies:

```bash
# Start services with dependencies
docker-compose up -d postgres redis
make run-all

# Or use full Docker setup
docker-compose up -d
```

### Azure Production

1. **Infrastructure Setup**
   ```bash
   # Deploy Azure resources
   az deployment group create \
     --resource-group rg-liveops \
     --template-file infrastructure/azure-resources.json
   ```

2. **Container Registry**
   ```bash
   # Build and push images
   make docker-build
   make docker-push
   ```

3. **Kubernetes Deployment**
   ```bash
   # Deploy to AKS
   kubectl apply -f k8s/
   ```

4. **Environment Configuration**
   - Update `.env` with Azure service endpoints
   - Configure Azure Key Vault for secrets
   - Set up Azure Monitor for observability

### CI/CD

GitHub Actions workflows are included for:
- **Build & Test**: On every push/PR
- **Security Scan**: CodeQL analysis
- **Deploy to Azure**: On main branch merge
- **Container Build**: Multi-stage Docker builds

## 📈 Monitoring & Observability

### Metrics

The system exposes Prometheus metrics:

- **Business Metrics**: Events processed, players active, rewards claimed
- **Technical Metrics**: Request latency, error rates, resource usage
- **Custom Metrics**: Fairness violations, leaderboard updates

Access metrics at: `http://localhost:9090/metrics`

### Logging

Structured JSON logging with configurable levels:

```bash
# View logs in development
make logs

# Production log aggregation (Azure Monitor)
az monitor log-analytics query \
  --workspace <workspace-id> \
  --analytics-query "ContainerLog | where Image contains 'liveops'"
```

### Health Checks

Each service exposes health endpoints:

- `/health` - Basic health check
- `/health/ready` - Readiness probe
- `/health/live` - Liveness probe

## 🔐 Security

### Authentication & Authorization

- **HMAC Signatures**: All API calls require HMAC-SHA256 signatures
- **Idempotency**: Duplicate request protection via idempotency keys
- **Rate Limiting**: Configurable rate limits per endpoint

### Data Protection

- **Encryption**: TLS 1.3 for all communications
- **Secrets Management**: Azure Key Vault integration
- **Audit Logging**: All events are logged for compliance

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and `gofmt` formatting
- Write tests for new features (minimum 80% coverage)
- Update documentation for API changes
- Use conventional commit messages

## 📋 Roadmap

- [ ] **Event Streaming**: Kafka/EventHub integration
- [ ] **Advanced Analytics**: ML-based anomaly detection
- [ ] **Multi-tenant**: Support for multiple games
- [ ] **GraphQL API**: Alternative to REST endpoints
- [ ] **Mobile SDK**: Client libraries for game integration
- [ ] **A/B Testing**: Feature flag management

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**tkdals69**
- GitHub: [@tkdals69](https://github.com/tkdals69)
- Project Link: [https://github.com/tkdals69/go-microservices](https://github.com/tkdals69/go-microservices)

## 🙏 Acknowledgments

- [Chi Router](https://github.com/go-chi/chi) - Lightweight HTTP router
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) - Azure integrations
- [go-redis](https://github.com/redis/go-redis) - Redis client
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver
- [Prometheus](https://prometheus.io/) - Monitoring system

---

⭐ **Star this repository if it helped you!** ⭐

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