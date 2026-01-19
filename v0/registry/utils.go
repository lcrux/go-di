package registry

import (
	"log"
	"os"
)

var isDebugEnabled = os.Getenv("GODI_DEBUG") == "true"

func debugLog(format string, v ...interface{}) {
	if isDebugEnabled {
		log.Printf(format, v...)
	}
}
