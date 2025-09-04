# Azure 배포를 위한 설정 및 스크립트

## 1. Azure Container Registry 생성

```bash
# 리소스 그룹 생성
az group create --name go-microservices-rg --location koreacentral

# Azure Container Registry 생성
az acr create \
  --resource-group go-microservices-rg \
  --name gomicroservicesacr \
  --sku Basic \
  --location koreacentral

# ACR 로그인
az acr login --name gomicroservicesacr
```

## 2. 이미지 빌드 및 푸시

```bash
# Gateway 서비스 빌드 및 푸시
docker build -f Dockerfile.gateway -t gomicroservicesacr.azurecr.io/gateway:v1.0 .
docker push gomicroservicesacr.azurecr.io/gateway:v1.0

# Web 서비스 빌드 및 푸시
docker build -f Dockerfile.web -t gomicroservicesacr.azurecr.io/web:v1.0 .
docker push gomicroservicesacr.azurecr.io/web:v1.0
```

## 3. Azure Container Instances 배포

### Gateway 서비스 배포
```bash
az container create \
  --resource-group go-microservices-rg \
  --name gateway-service \
  --image gomicroservicesacr.azurecr.io/gateway:v1.0 \
  --registry-login-server gomicroservicesacr.azurecr.io \
  --registry-username gomicroservicesacr \
  --registry-password $(az acr credential show --name gomicroservicesacr --query "passwords[0].value" --output tsv) \
  --dns-name-label gomicroservices-gateway \
  --ports 8085 \
  --environment-variables \
    PORT=8085 \
    CLOUD=azure \
    BUS_KIND=inmem \
    HMAC_SECRET=azure-secret-key-32bytes-minimum-length-required \
  --cpu 1 \
  --memory 1.5 \
  --location koreacentral
```

### Web 서비스 배포
```bash
az container create \
  --resource-group go-microservices-rg \
  --name web-service \
  --image gomicroservicesacr.azurecr.io/web:v1.0 \
  --registry-login-server gomicroservicesacr.azurecr.io \
  --registry-username gomicroservicesacr \
  --registry-password $(az acr credential show --name gomicroservicesacr --query "passwords[0].value" --output tsv) \
  --dns-name-label gomicroservices-web \
  --ports 3003 \
  --environment-variables \
    WEB_PORT=3003 \
  --cpu 1 \
  --memory 1.5 \
  --location koreacentral
```

## 4. 배포 확인

### 서비스 상태 확인
```bash
# Gateway 서비스 상태
az container show --resource-group go-microservices-rg --name gateway-service --query "instanceView.state" --output table

# Web 서비스 상태  
az container show --resource-group go-microservices-rg --name web-service --query "instanceView.state" --output table
```

### 접속 URL 확인
```bash
# Gateway 접속 URL
echo "Gateway URL: http://$(az container show --resource-group go-microservices-rg --name gateway-service --query "ipAddress.fqdn" --output tsv):8085"

# Web 접속 URL
echo "Web URL: http://$(az container show --resource-group go-microservices-rg --name web-service --query "ipAddress.fqdn" --output tsv):3003"
```

## 5. 로그 확인

```bash
# Gateway 서비스 로그
az container logs --resource-group go-microservices-rg --name gateway-service

# Web 서비스 로그
az container logs --resource-group go-microservices-rg --name web-service
```

## 6. 정리 (필요시)

```bash
# 리소스 그룹 삭제 (모든 리소스 삭제됨)
az group delete --name go-microservices-rg --yes --no-wait
```

## 예상 비용 (월 기준)
- Azure Container Registry (Basic): ~$5
- Azure Container Instances (2개 인스턴스): ~$30-40
- 네트워크 트래픽: ~$5-10
- **총 예상 비용: $40-55/월**
