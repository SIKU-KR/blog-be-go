package utils

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

// ConvertToWebP는 이미지를 WebP 형식으로 변환합니다.
func ConvertToWebP(fileBytes []byte) ([]byte, error) {
	// libvips를 사용하여 이미지 변환
	image := bimg.NewImage(fileBytes)

	// 이미지 크기 최적화 옵션 설정
	options := bimg.Options{
		Type:    bimg.WEBP, // WebP 포맷으로 변환
		Quality: 80,        // 품질 (1-100)
	}

	// 이미지 변환 실행
	return image.Process(options)
}

// GenerateUniqueFileName은 중복 없는 파일명을 생성합니다.
func GenerateUniqueFileName(originalName string) string {
	fileExt := ".webp" // WebP로 변환하기 때문에 확장자는 항상 .webp
	uniqueID := uuid.New().String()
	return uniqueID + fileExt
}

// UploadToS3는 변환된 이미지를 S3에 업로드합니다.
func UploadToS3(ctx context.Context, s3Client *s3.Client, fileContent []byte, fileName string) (string, error) {
	bucketName := os.Getenv("S3_BUCKET_NAME")

	// 이미지를 버킷 루트에 직접 저장
	objectKey := fileName

	// S3에 업로드
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(fileContent),
		ContentType: aws.String("image/webp"),
	})

	if err != nil {
		return "", err
	}

	// S3 URL 생성
	s3Region := os.Getenv("AWS_REGION")
	if s3Region == "" {
		s3Region = "ap-northeast-2" // 기본값
	}

	s3URL := "https://" + bucketName + ".s3." + s3Region + ".amazonaws.com/" + objectKey

	return s3URL, nil
}

// ProcessImage는 이미지 파일을 처리하고 S3에 업로드합니다.
func ProcessImage(ctx context.Context, s3Client *s3.Client, file *multipart.FileHeader) ([]byte, string, string, error) {
	// 파일 열기
	src, err := file.Open()
	if err != nil {
		return nil, "", "", err
	}
	defer src.Close()

	// 파일 내용 읽기
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, "", "", err
	}

	// WebP로 변환
	webpBytes, err := ConvertToWebP(fileBytes)
	if err != nil {
		return nil, "", "", err
	}

	// 고유한 파일명 생성
	fileName := GenerateUniqueFileName(file.Filename)

	// S3에 업로드
	s3URL, err := UploadToS3(ctx, s3Client, webpBytes, fileName)
	if err != nil {
		return nil, "", "", err
	}

	return webpBytes, fileName, s3URL, nil
}
