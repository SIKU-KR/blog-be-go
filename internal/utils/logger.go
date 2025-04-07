package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

const (
	// LogLevels
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
	LogLevelDebug = "DEBUG"
)

// Logger CloudWatch 로깅 유틸리티
type Logger struct {
	client          *cloudwatchlogs.Client
	logGroupName    string
	logStreamPrefix string
	logStreamName   string
	sequenceToken   *string
	env             string
}

// NewLogger Logger 인스턴스 생성
func NewLogger(client *cloudwatchlogs.Client) *Logger {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	logGroupName := os.Getenv("CLOUDWATCH_LOG_GROUP")
	if logGroupName == "" {
		logGroupName = "bumsiku-api"
	}

	// 로그 스트림 이름: {prefix}-{timestamp}
	timestamp := time.Now().Format("2006-01-02")
	logStreamName := fmt.Sprintf("%s-%s", env, timestamp)

	logger := &Logger{
		client:        client,
		logGroupName:  logGroupName,
		logStreamName: logStreamName,
		env:           env,
	}

	// 로그 그룹과 스트림 초기화
	_ = logger.initLogGroupAndStream(context.Background())

	return logger
}

// initLogGroupAndStream 로그 그룹과 스트림을 초기화합니다.
func (l *Logger) initLogGroupAndStream(ctx context.Context) error {
	// 로그 그룹 생성 (이미 존재해도 오류 발생하지 않음)
	_, err := l.client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(l.logGroupName),
	})
	if err != nil {
		// 이미 존재하는 로그 그룹이면 무시
		fmt.Printf("로그 그룹 생성 중 알림: %v\n", err)
	}

	// 로그 스트림 생성 (이미 존재해도 오류 발생하지 않음)
	_, err = l.client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(l.logGroupName),
		LogStreamName: aws.String(l.logStreamName),
	})
	if err != nil {
		// 이미 존재하는 로그 스트림이면 무시
		fmt.Printf("로그 스트림 생성 중 알림: %v\n", err)
	}

	return nil
}

// Log 지정된 로그 레벨로 로그를 남깁니다
func (l *Logger) Log(ctx context.Context, level, message string, fields map[string]string) error {
	// 로그 엔트리에 환경 정보 추가
	if fields == nil {
		fields = make(map[string]string)
	}
	fields["env"] = l.env

	// 필드를 문자열로 변환
	fieldStr := ""
	for k, v := range fields {
		fieldStr += fmt.Sprintf(" %s=%s", k, v)
	}

	logEvent := fmt.Sprintf("[%s]%s %s", level, fieldStr, message)

	// 현재 타임스탬프
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// 로그 이벤트 전송
	input := &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(l.logGroupName),
		LogStreamName: aws.String(l.logStreamName),
		LogEvents: []types.InputLogEvent{
			{
				Message:   aws.String(logEvent),
				Timestamp: aws.Int64(timestamp),
			},
		},
	}

	// 시퀀스 토큰이 있으면 추가
	if l.sequenceToken != nil {
		input.SequenceToken = l.sequenceToken
	}

	// 로그 전송
	resp, err := l.client.PutLogEvents(ctx, input)
	if err != nil {
		fmt.Printf("로그 전송 실패: %v\n", err)
		return err
	}

	// 다음 시퀀스 토큰 저장
	l.sequenceToken = resp.NextSequenceToken

	return nil
}

// Info 정보 레벨 로그
func (l *Logger) Info(ctx context.Context, message string, fields map[string]string) error {
	return l.Log(ctx, LogLevelInfo, message, fields)
}

// Warn 경고 레벨 로그
func (l *Logger) Warn(ctx context.Context, message string, fields map[string]string) error {
	return l.Log(ctx, LogLevelWarn, message, fields)
}

// Error 에러 레벨 로그
func (l *Logger) Error(ctx context.Context, message string, fields map[string]string) error {
	return l.Log(ctx, LogLevelError, message, fields)
}

// Debug 디버그 레벨 로그
func (l *Logger) Debug(ctx context.Context, message string, fields map[string]string) error {
	return l.Log(ctx, LogLevelDebug, message, fields)
}
