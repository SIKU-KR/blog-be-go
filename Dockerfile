# 빌드 단계: Golang 1.23-alpine 이미지를 사용하여 테스트 및 빌드 수행
FROM golang:1.23-alpine AS builder
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

# ARM 아키텍쳐용 바이너리 빌드 (CGO 비활성화)
RUN env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o serverapp .

# 실행 단계: ARM 아키텍쳐용 경량 알파인 리눅스 이미지 사용
FROM --platform=linux/arm64 alpine:latest
WORKDIR /app

# 빌드 단계에서 생성된 serverapp 바이너리 복사
COPY --from=builder /app/serverapp .

# 애플리케이션이 사용하는 포트 (문서화 목적)
EXPOSE 8080

# 서버 애플리케이션 실행
CMD ["./serverapp"]
