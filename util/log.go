package util

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	//"time"
)

// DEBUG => DEBUG, INFO, ERROR
// INFO => INFO, ERROR
// ERROR => ERROR
func IsLogged(logMode string) bool {
	logModeEnv := os.Getenv("LOG_MODE")

	// ERROR
	if logModeEnv == "ERROR" {
		return logMode == logModeEnv
	}
	// INFO
	if logModeEnv == "INFO" {
		if logMode == logModeEnv {
			return true
		}
		if logMode == "ERROR" {
			return true
		}
		return false
	}
	// DEBUG: will return true
	return true
}

func Log(logMode string, packageName string, funcName string, message string) {
	if !IsLogged(logMode) {
		return
	}
	logEntry := fmt.Sprintf("[%s] => %s > %s > %s", logMode, packageName, funcName, message)
	log.Println(logEntry)
}

func LogError(packageName string, funcName string, message string, err error) {
	if !IsLogged("ERROR") {
		return
	}
	logEntry := fmt.Sprintf("[ERROR] => %s > %s > %s", packageName, funcName, message)
	if err != nil {
		logEntry += fmt.Sprintf("\nError: %v\nStack Trace: %s", err, debug.Stack())
	}
	log.Println(logEntry)
}

func LogAPI(clientIP string, targetAPI string, statusCode int, elapsedTime string) {
	if !IsLogged("INFO") {
		return
	}
	logMessage := fmt.Sprintf("[INFO] => API > IP: %s, requested path: %s, status: %d, elapsed time: %s", clientIP, targetAPI, statusCode, elapsedTime)
	log.Println(logMessage)
}
