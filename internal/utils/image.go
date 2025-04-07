package utils

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// ConvertToOptimizedImage는 이미지를 최적화하여 JPEG로 변환합니다.
func ConvertToOptimizedImage(fileBytes []byte) ([]byte, error) {
	// 원본 이미지 디코딩
	src, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}

	// 이미지 리사이징 (필요한 경우)
	// 여기서는 원본 크기를 유지하면서 품질만 조정
	img := imaging.Resize(src, 0, 0, imaging.Lanczos)

	// 최적화된 이미지를 버퍼에 저장
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateUniqueFileName은 중복 없는 파일명을 생성합니다.
func GenerateUniqueFileName(originalName string) string {
	fileExt := ".jpg" // JPEG로 변환하기 때문에 확장자는 항상 .jpg
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
		ContentType: aws.String("image/jpeg"),
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

	// 이미지 최적화
	processedBytes, err := ConvertToOptimizedImage(fileBytes)
	if err != nil {
		return nil, "", "", err
	}

	// 고유한 파일명 생성
	fileName := GenerateUniqueFileName(file.Filename)

	// S3에 업로드
	s3URL, err := UploadToS3(ctx, s3Client, processedBytes, fileName)
	if err != nil {
		return nil, "", "", err
	}

	return processedBytes, fileName, s3URL, nil
}
