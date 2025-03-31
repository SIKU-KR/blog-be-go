# AWS 클라이언트 사용 가이드

## 초기화

AWS 클라이언트는 DynamoDB와 S3 각각 독립적으로 초기화할 수 있습니다:

### DynamoDB 클라이언트 초기화
```go
ctx := context.Background()
ddbClient, err := client.NewDdbClient(ctx)
if err != nil {
    log.Fatalf("DynamoDB 클라이언트 초기화 실패: %v", err)
}
```

### S3 클라이언트 초기화
```go
ctx := context.Background()
s3Client, err := client.NewS3Client(ctx)
if err != nil {
    log.Fatalf("S3 클라이언트 초기화 실패: %v", err)
}
```

## 자격 증명 설정

AWS 서비스를 사용하기 위해서는 다음 중 하나의 방법으로 자격 증명을 설정해야 합니다:

### 1. 환경 변수 사용
```bash
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_REGION=ap-northeast-2  # 예: 서울 리전
```

### 2. AWS 자격 증명 파일 사용
`~/.aws/credentials` 파일에 다음과 같이 설정:
```ini
[default]
aws_access_key_id = your_access_key
aws_secret_access_key = your_secret_key
```

`~/.aws/config` 파일에 리전 설정:
```ini
[default]
region = ap-northeast-2
```

### 3. IAM 역할 사용
AWS 환경(EC2, ECS, Lambda 등)에서 실행 시 IAM 역할이 자동으로 사용됩니다.

## S3 사용 예시

### 버킷 목록 조회
```go
output, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
if err != nil {
    log.Printf("버킷 목록 조회 실패: %v", err)
    return
}

for _, bucket := range output.Buckets {
    fmt.Printf("버킷 이름: %s, 생성일: %s\n", *bucket.Name, bucket.CreationDate)
}
```

### 파일 업로드
```go
file, err := os.Open("example.jpg")
if err != nil {
    log.Printf("파일 열기 실패: %v", err)
    return
}
defer file.Close()

input := &s3.PutObjectInput{
    Bucket: aws.String("your-bucket-name"),
    Key:    aws.String("uploads/example.jpg"),
    Body:   file,
}

_, err = s3Client.PutObject(ctx, input)
if err != nil {
    log.Printf("파일 업로드 실패: %v", err)
    return
}
```

## DynamoDB 사용 예시

### 테이블 목록 조회
```go
output, err := ddbClient.ListTables(ctx, &dynamodb.ListTablesInput{})
if err != nil {
    log.Printf("테이블 목록 조회 실패: %v", err)
    return
}

for _, tableName := range output.TableNames {
    fmt.Printf("테이블 이름: %s\n", tableName)
}
```

### 아이템 추가
```go
item := map[string]types.AttributeValue{
    "ID": &types.AttributeValueMemberS{Value: "123"},
    "Name": &types.AttributeValueMemberS{Value: "테스트 아이템"},
    "CreatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
}

input := &dynamodb.PutItemInput{
    TableName: aws.String("your-table-name"),
    Item:     item,
}

_, err = ddbClient.PutItem(ctx, input)
if err != nil {
    log.Printf("아이템 추가 실패: %v", err)
    return
}
```

### 아이템 조회
```go
key := map[string]types.AttributeValue{
    "ID": &types.AttributeValueMemberS{Value: "123"},
}

input := &dynamodb.GetItemInput{
    TableName: aws.String("your-table-name"),
    Key:       key,
}

result, err := ddbClient.GetItem(ctx, input)
if err != nil {
    log.Printf("아이템 조회 실패: %v", err)
    return
}

// 결과 처리
if result.Item == nil {
    fmt.Println("아이템을 찾을 수 없습니다")
    return
}

fmt.Printf("조회된 아이템: %v\n", result.Item)
```

## 주의사항

1. 항상 context를 적절히 관리하여 리소스 누수를 방지하세요.
2. 에러 처리를 항상 수행하세요.
3. 민감한 자격 증명 정보는 환경 변수나 안전한 시크릿 관리 서비스를 통해 관리하세요.
4. 대용량 파일 처리 시 메모리 사용량을 고려하여 스트리밍 방식을 사용하세요.
5. 각 클라이언트는 독립적으로 생성되므로, 필요한 서비스의 클라이언트만 초기화하여 사용하세요. 