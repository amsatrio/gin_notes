package initializer

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func LoadEnvironmentVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("error: %v\n", err)
		log.Fatal("Error loading .env file")
		return
	}

	mode := os.Getenv("GIN_MODE")
	gin.SetMode(mode)
}
