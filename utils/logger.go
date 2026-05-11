package utils

import (
	"fmt"
	"time"
)

func log(level string, msg string) {
	fmt.Printf("[%s] [%s] %s\n", level, time.Now().Format("15:04:05"), msg)
}

func Info(msg string) {
	log("INFO", msg)
}

func Success(msg string) {
	log("SUCCESS", msg)
}

func Error(msg string) {
	log("ERROR", msg)
}
