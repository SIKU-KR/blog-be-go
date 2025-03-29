package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// 개발 전용. 운영환경에서는 시스템 변수 사용
func LoadEnv() error {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("failed to load go.env: %v", err)
			return err
		}
		log.Println(".env file loaded")
	} else {
		log.Println(".env file not found.")
	}
	return nil
}
