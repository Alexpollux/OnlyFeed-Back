package logs

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// The different severities and their meaning :
// - INFO = request successful
// - WARN = not authorized or not found for example (minor error)
// - ERROR = error occurred due to a problem
// - FATAL = error completely blocking the app

var logger = log.New(os.Stdout, "", 0)

func LogJSON(level, message string, fields map[string]interface{}) {
	logEntry := map[string]interface{}{
		"severity": level, // "DEBUG", "INFO", "WARN", "ERROR" & "FATAL"
		"message":  message,
		"time":     time.Now().Format(time.RFC3339),
	}
	for k, v := range fields {
		logEntry[k] = v
	}
	jsonLog, _ := json.Marshal(logEntry)
	logger.Println(string(jsonLog))
}
