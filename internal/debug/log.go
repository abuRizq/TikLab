package debug

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const logFile = "debug-06bbec.log"

var mu sync.Mutex

// Log writes an NDJSON line to the debug log file.
func Log(location, message string, data map[string]interface{}, hypothesisId string) {
	mu.Lock()
	defer mu.Unlock()
	path := logFile
	if wd, err := os.Getwd(); err == nil {
		path = filepath.Join(wd, logFile)
	}
	entry := map[string]interface{}{
		"sessionId":    "06bbec",
		"location":     location,
		"message":      message,
		"data":         data,
		"hypothesisId": hypothesisId,
		"timestamp":    time.Now().UnixMilli(),
	}
	line, _ := json.Marshal(entry)
	line = append(line, '\n')
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(line)
}
