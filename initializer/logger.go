package initializer

import (
	"fmt"
	"os"
	//"io"
	"log"
	"time"

	//"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/util"
)

func LoggerInit() {
	// clear log
	if os.Getenv("LOG_CLEAR") == "true" {
		err := util.RemoveAll("log")
		if err != nil {
			fmt.Println("delete log error: " + err.Error())
		}
	}

	requestTime := time.Now().Format("2006-01-02")

	// Create a log directory if it doesn't exist
	if _, err := os.Stat("log"); os.IsNotExist(err) {
		err := os.Mkdir("log", os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create log directory")
			return
		}
	}

	// Create a log file if it doesn't exist
	logFileName := "log/log_" + requestTime + ".log"
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		defer func(logFile *os.File) {
			err := logFile.Close()
			if err != nil {
				log.Println(err)
			}
		}(logFile)
		log.Fatal("Failed to create or open log file:", err)
		return
	}

	// Set the log file as the default writer for Gin
	//gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

	// Set the log file as the output for the standard logger
	log.SetOutput(logFile)

	// ASCII
	asciiArt := `
    ____
   < hi there >
    ----
         \   ^__^
          \  (oo)\_______
             (__)\       )\/\
                 ||----w |
                 ||     ||
    `

	log.Println(asciiArt)
}
