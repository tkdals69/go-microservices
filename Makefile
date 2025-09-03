build:
    go build -o bin/gateway ./cmd/gateway
    go build -o bin/progression ./cmd/progression
    go build -o bin/leaderboard ./cmd/leaderboard
    go build -o bin/fairness ./cmd/fairness

run-gateway:
    go run ./cmd/gateway

run-progression:
    go run ./cmd/progression

run-leaderboard:
    go run ./cmd/leaderboard

run-fairness:
    go run ./cmd/fairness

test:
    go test ./pkg/...