package model

// UploadImageRequest는 이미지 업로드 요청 데이터를 담는 구조체입니다.
type UploadImageRequest struct {
	// 이미지 파일이 멀티파트 폼으로 전송됩니다
}

// UploadImageResponse는 이미지 업로드 응답 데이터를 담는 구조체입니다.
type UploadImageResponse struct {
	URL       string `json:"url" example:"https://bumsiku-bucket.s3.ap-northeast-2.amazonaws.com/image-uuid.webp"` // 업로드된 이미지 URL
	FileName  string `json:"fileName" example:"image-uuid.webp"`                                                   // 업로드된 이미지 파일명
	Size      int64  `json:"size" example:"102400"`                                                                // 이미지 크기 (바이트)
	MimeType  string `json:"mimeType" example:"image/webp"`                                                        // MIME 타입
	Timestamp int64  `json:"timestamp" example:"1617235200"`                                                       // 업로드 시간 (Unix timestamp)
}
