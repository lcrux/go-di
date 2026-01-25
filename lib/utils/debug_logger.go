package utils

import (
	"log"
	"os"
)

func DebugLog(format string, v ...interface{}) {
	if os.Getenv("GODI_DEBUG") == "true" {
		log.Printf(format, v...)
	}
}
