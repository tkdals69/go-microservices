package config

import (
    "log"
    "os"
    "strings"
    "time"
    "strconv"
)

type Config struct {
    Cloud        string
    Env          string
    Port         string
    HMACSecret   string
    DBUrl        string
    RedisUrl     string
    BusKind      string
    BusConn      string
    LBWindows    []string
    LBTopN       int
    SnapshotCron string
    AnomalyZ     float64
    BlockSec     int
    RateLimitMax int
    IdempTTL     time.Duration
}

func Load() *Config {
    cloud := strings.ToLower(os.Getenv("CLOUD"))
    if cloud != "azure" {
        log.Printf("[WARN] CLOUD 값이 잘못됨: %s, azure만 지원", cloud)
        cloud = "azure"
    }
    env := os.Getenv("ENV")
    port := os.Getenv("PORT")
    hmacSecret := os.Getenv("HMAC_SECRET")
    dbUrl := os.Getenv("DB_URL")
    redisUrl := os.Getenv("REDIS_URL")

    if hmacSecret == "" || len(hmacSecret) < 32 {
        log.Fatal("HMAC_SECRET 미설정 또는 32바이트 미만, 프로세스 종료")
    }
    if dbUrl == "" {
        log.Fatal("DB_URL 미설정, 프로세스 종료")
    }
    if redisUrl == "" {
        log.Fatal("REDIS_URL 미설정, 프로세스 종료")
    }
    if strings.HasPrefix(redisUrl, "rediss://") && !strings.Contains(redisUrl, ":6380") {
        log.Printf("[WARN] Azure Redis는 TLS(6380) 필수, 현재: %s", redisUrl)
    }

    busKind := os.Getenv("BUS_KIND")
    busConn := os.Getenv("BUS_CONN")
    lbWindows := []string{"daily", "weekly", "seasonal"}
    if v := os.Getenv("LB_WINDOWS"); v != "" {
        lbWindows = strings.Split(v, ",")
    }
    lbTopN := 100
    if v := os.Getenv("LB_TOPN"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            lbTopN = n
        }
    }
    snapshotCron := os.Getenv("SNAPSHOT_CRON")
    if snapshotCron == "" {
        snapshotCron = "0 */15 * * * *"
    }
    anomalyZ := 3.5
    if v := os.Getenv("ANOMALY_Z"); v != "" {
        if f, err := strconv.ParseFloat(v, 64); err == nil {
            anomalyZ = f
        }
    }
    blockSec := 60
    if v := os.Getenv("BLOCK_SEC"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            blockSec = n
        }
    }
    rateLimitMax := 50
    if v := os.Getenv("RATE_LIMIT_MAX"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            rateLimitMax = n
        }
    }
    idempTTL := 300 * time.Second
    if v := os.Getenv("IDEMP_TTL"); v != "" {
        if d, err := time.ParseDuration(v); err == nil {
            idempTTL = d
        }
    }

    return &Config{
        Cloud:        cloud,
        Env:          env,
        Port:         port,
        HMACSecret:   hmacSecret,
        DBUrl:        dbUrl,
        RedisUrl:     redisUrl,
        BusKind:      busKind,
        BusConn:      busConn,
        LBWindows:    lbWindows,
        LBTopN:       lbTopN,
        SnapshotCron: snapshotCron,
        AnomalyZ:     anomalyZ,
        BlockSec:     blockSec,
        RateLimitMax: rateLimitMax,
        IdempTTL:     idempTTL,
    }
}