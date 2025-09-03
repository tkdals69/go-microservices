package observability

import (
	"encoding/json"
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log("INFO", message, fields)
}

func (l *Logger) Error(message string, fields map[string]interface{}) {
	l.log("ERROR", message, fields)
}

func (l *Logger) log(level string, message string, fields map[string]interface{}) {
	logEntry := map[string]interface{}{
		"level":   level,
		"message": message,
	}

	for key, value := range fields {
		logEntry[key] = value
	}

	jsonEntry, err := json.Marshal(logEntry)
	if err != nil {
		l.Logger.Println("Failed to marshal log entry:", err)
		return
	}

	l.Logger.Println(string(jsonEntry))
}