package logs

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var logger = log.New(os.Stdout, "", 0)

func LogJSON(level, message string, route string, fields map[string]interface{}) {
	logEntry := map[string]interface{}{
		"severity": level, // "INFO", "ERROR", etc.
		"message":  message,
		"route":    route,
		"time":     time.Now().Format(time.RFC3339),
	}
	for k, v := range fields {
		logEntry[k] = v
	}
	jsonLog, _ := json.Marshal(logEntry)
	logger.Println(string(jsonLog))
}
