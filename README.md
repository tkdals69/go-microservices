# LiveOps Progression & Fairness Guard

- Author: tkdals69
- Cloud: Azure (CLOUD=azure)
- Go 1.22, chi, pgx, go-redis, prometheus

## 환경변수(.env 예시)
```env
CLOUD=azure
ENV=dev
PORT=8080
HMAC_SECRET=change-me-32bytes-min
DB_URL=postgres://liveops:STRONG_PASS@<pg-name>.postgres.database.azure.com:5432/liveops?sslmode=require
REDIS_URL=rediss://:PRIMARY_KEY@<name>.redis.cache.windows.net:6380/0
...
```

## 실행
```bash
go mod tidy
make build
make run-gateway
make run-progression
make run-leaderboard
make run-fairness
```

## 테스트
```bash
make test
```

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