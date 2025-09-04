# 실제 구축된 시스템 기반 인프라 계획

## 🎯 현재 시스템 현황
- **Gateway 서비스**: 8085 포트, IP 기반 플레이어 구분, 인메모리 리더보드
- **웹 클라이언트**: 3003 포트, 클릭 게임 인터페이스
- **저장소**: 인메모리 (Redis/DB 없음)
- **아키텍처**: 모놀리틱에 가까운 단일 서비스

## 📋 단계별 실용적 인프라 구축 계획

### Phase 1: 현재 시스템 컨테이너화 및 클라우드 배포
**목표**: 현재 동작하는 시스템을 Azure에 배포

#### 1.1 Docker 컨테이너화
```dockerfile
# Dockerfile.gateway
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gateway ./cmd/gateway

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gateway .
EXPOSE 8085
CMD ["./gateway"]
```

```dockerfile
# Dockerfile.web
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o web ./web

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/web .
COPY --from=builder /app/web/static ./static
COPY --from=builder /app/web/templates ./templates
EXPOSE 3003
CMD ["./web"]
```

#### 1.2 Azure Container Instances 배포
- 빠른 배포를 위해 ACI 사용 (Kubernetes보다 단순)
- 환경변수로 설정 주입
- Azure Container Registry 사용

#### 1.3 기본 모니터링
- Application Insights 연동
- 기본 로깅 및 메트릭 수집

### Phase 2: 데이터 지속성 추가
**목표**: 인메모리 저장소를 실제 데이터베이스로 교체

#### 2.1 Azure Database for PostgreSQL 연동
- 현재 인메모리 리더보드를 PostgreSQL로 마이그레이션
- 연결 풀링 및 트랜잭션 처리

#### 2.2 Azure Cache for Redis 추가
- 실시간 리더보드용 캐시 레이어
- 세션 관리 (필요시)

#### 2.3 데이터 백업 전략
- 자동 백업 설정
- Point-in-time recovery 구성

### Phase 3: 확장성 및 고가용성
**목표**: 트래픽 증가에 대응할 수 있는 구조로 발전

#### 3.1 Azure App Service 또는 AKS 마이그레이션
- 현재 시스템을 스케일 가능한 플랫폼으로 이전
- 로드 밸런싱 구성

#### 3.2 CDN 및 정적 자산 최적화
- Azure CDN으로 웹 자산 배포
- 이미지 및 정적 파일 최적화

#### 3.3 API Gateway 추가
- Azure API Management 도입
- Rate limiting, 인증 등 추가

### Phase 4: 마이크로서비스 분리 (선택사항)
**목표**: 필요시 서비스 분리

#### 4.1 서비스 분리 전략
- Leaderboard 서비스 분리
- Progression 서비스 분리
- Event processing 서비스 분리

#### 4.2 Service Mesh 도입
- Istio 또는 Linkerd 고려
- 서비스 간 통신 관리

## 💰 예상 비용 (월 기준)
### Phase 1 (기본 배포)
- Azure Container Registry: $5
- Azure Container Instances (2개): $30-50
- Azure Application Insights: $10-20
- **총 예상: $45-75/월**

### Phase 2 (데이터 지속성)
- Azure Database for PostgreSQL (Basic): $20-40
- Azure Cache for Redis (Basic): $15-25
- 스토리지: $5-10
- **총 예상: $85-150/월**

### Phase 3 (확장성)
- Azure App Service 또는 AKS: $50-150
- Azure CDN: $10-20
- API Management: $50-100
- **총 예상: $195-420/월**

## 🚀 즉시 실행 가능한 다음 단계

### 1주차: 컨테이너화
- [ ] Dockerfile 작성
- [ ] Docker Compose로 로컬 테스트
- [ ] Azure Container Registry 설정

### 2주차: 클라우드 배포
- [ ] Azure Container Instances 배포
- [ ] 환경변수 및 시크릿 관리
- [ ] 도메인 연결 (선택사항)

### 3주차: 모니터링 및 로깅
- [ ] Application Insights 연동
- [ ] 대시보드 구성
- [ ] 알림 설정

## 📝 주요 고려사항

### 현실적 접근
1. **과도한 마이크로서비스화 지양**: 현재 단일 서비스로 충분
2. **점진적 확장**: 트래픽과 요구사항에 따라 단계적 발전
3. **비용 효율성**: 초기에는 관리형 서비스 활용으로 운영 부담 최소화

### 운영 우선순위
1. **가용성 확보**: 기본적인 고가용성 구성
2. **모니터링 강화**: 장애 감지 및 대응 체계
3. **백업 및 복구**: 데이터 손실 방지 전략

이 계획은 현재 구축된 시스템의 실제 상태를 반영하여, 과도한 설계보다는 실용적이고 단계적인 접근을 제안합니다.
