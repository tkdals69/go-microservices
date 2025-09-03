module github.com/tkdals69/go-microservices

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.7
    github.com/go-chi/render v1.0.3
    github.com/redis/go-redis/v9 v9.0.5
    github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.2.0
    github.com/prometheus/client_golang v1.11.0
    github.com/joho/godotenv v1.4.0
    github.com/jackc/pgx/v5 v5.5.4
    github.com/jmoiron/sqlx v1.3.5
)

// 보안 취약점 패치를 위한 간접 의존성 업그레이드
replace (
    github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.17.0
)