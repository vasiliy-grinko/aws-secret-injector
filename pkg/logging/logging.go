package logging

import (
	"fmt"
	"io"
	"log"
	"os"

	"swisscom/config"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	histLength = 10
)

var (
	Info             *log.Logger
	Error            *log.Logger
	Debug            *log.Logger
	FileLog          *log.Logger
	LumberjackLogger *lumberjack.Logger
	errorHistory     []string
	syncOnce         = new(sync.Once)
)

func InitLogger(cfg *config.Config) error {
	syncOnce.Do(func() {
		LumberjackLogger = &lumberjack.Logger{
			Filename:   "/etc/tosser/logs/log.txt",
			MaxSize:    50, // megabytes
			MaxBackups: 5,
			MaxAge:     30, //days
			LocalTime:  true,
		}
	})

	if err := LumberjackLogger.Rotate(); err != nil {
		return fmt.Errorf("failed to open log file: %s", err)
	}

	multi := io.MultiWriter(LumberjackLogger, os.Stdout)
	Debug = log.New(multi, "DEBUG: ", log.Ldate|log.Ltime)
	Info = log.New(multi, "INFO:  ", log.Ldate|log.Ltime)

	// multi := io.MultiWriter(LumberjackLogger, os.Stderr)
	Error = log.New(multi, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	file, err := os.OpenFile("/etc/tosser/logs/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log.txt file: %s", err)
	}
	FileLog = log.New(file, "", log.Ldate|log.Ltime)
	return nil
}

func Errorln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	Error.Print(s)
	SaveErrorHistory(s)
}

func Errorf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	Error.Println(s)
	SaveErrorHistory(s)
}

func SaveErrorHistory(s string) {
	tm := time.Now().Format("2006-01-02 15:04:05")
	errorHistory = append(errorHistory, fmt.Sprintf("%s %s", tm, s))
	if len(errorHistory) > histLength {
		errorHistory = errorHistory[1:]
	}
}
