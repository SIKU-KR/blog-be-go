# 빌드 단계: Golang 1.20-alpine 이미지를 사용하여 테스트 및 빌드 수행
FROM golang:1.20-alpine AS builder
WORKDIR /app

# 테스트에 필요한 패키지 설치 (예: git)
RUN apk update && apk add --no-cache git

# 모듈 의존성 관리를 위해 go.mod와 go.sum 복사 후 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 애플리케이션 소스 전체 복사
COPY . .

# 모든 패키지에 대한 테스트 실행
RUN go test -v ./...

# x86_64 아키텍쳐용 바이너리 빌드 (t2.micro는 x86_64 아키텍쳐)
RUN env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o serverapp .

# 실행 단계: Amazon Linux 2023 이미지 사용
FROM --platform=linux/amd64 amazonlinux:2023
WORKDIR /app

# ca-certificates는 SSL/TLS 인증서 관련 패키지로, HTTPS 연결 시 필요합니다.
# 애플리케이션이 외부 HTTPS API를 호출하거나 TLS 통신이 필요한 경우 설치합니다.
RUN yum update -y && yum clean all

# 빌드 단계에서 생성된 serverapp 바이너리 복사
COPY --from=builder /app/serverapp .

# 애플리케이션이 사용하는 포트 (문서화 목적)
EXPOSE 8080

# 서버 애플리케이션 실행 (0.0.0.0:8080에서 리슨)
ENV HOST=0.0.0.0
ENV PORT=8080
CMD ["./serverapp"]
