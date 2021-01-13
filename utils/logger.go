package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/jackc/pgx/v4"
)

const (
	envKey        = "ENV"
	logFile       = "log/%s.log"
	skippedENV    = "test"
	logLevelKey   = "LOG_LEVEL"
	defProjFolder = "/fasthttp_template/"
	projFolderKey = "PROJECT_FOLDER"
	chmod         = 0666
)

var (
	levels = map[string]int{
		"notice":  1,
		"error":   2,
		"warning": 3,
		"info":    4,
		"debug":   5,
	}

	invLevels = make(map[int]string)
	logLevel  = levels["debug"]
)

// PgxLogger ...
type PgxLogger struct{}

// Init ...
func Init() {
	env := GetENV()

	envLevel := os.Getenv(logLevelKey)

	if len(envLevel) != 0 {
		logLevel = levels[envLevel]
	}

	for k, v := range levels {
		invLevels[v] = k
	}

	if env != skippedENV {
		logFile, err := os.OpenFile(
			fmt.Sprintf(logFile, env),
			os.O_CREATE|os.O_APPEND|os.O_RDWR,
			chmod,
		)

		if err != nil {
			panic(err)
		}

		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
	}

	LogNotice("Current log level: ", invLevels[logLevel])
}

// Log ...
func (l *PgxLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	if len(data) == 1 {
		return
	}

	json, _ := json.Marshal(data["args"])
	args := string(json)

	LogDebug(data["sql"])
	LogDebug(args)
	LogDebug(data["time"])
}

// LogInfo ...
func LogInfo(items ...interface{}) {
	if isLevelEnabled(levels["info"]) {
		log.Println(getLoggerLine(), PrefixInfo, fmt.Sprint(items...))
	}
}

// LogNotice ...
func LogNotice(items ...interface{}) {
	if isLevelEnabled(levels["notice"]) {
		log.Println(getLoggerLine(), PrefixNotice, fmt.Sprint(items...))
	}
}

// LogWarning ...
func LogWarning(items ...interface{}) {
	if isLevelEnabled(levels["warning"]) {
		log.Println(getLoggerLine(), PrefixWarning, fmt.Sprint(items...))
	}
}

// LogError ...
func LogError(items ...interface{}) {
	if isLevelEnabled(levels["error"]) {
		log.Println(getLoggerLine(), PrefixError, fmt.Sprint(items...))
	}
}

// LogDebug ...
func LogDebug(items ...interface{}) {
	if isLevelEnabled(levels["debug"]) {
		log.Println(getLoggerLine(), PrefixDebug, fmt.Sprint(items...))
	}
}

func isLevelEnabled(lvl int) bool {
	return (GetENV() != skippedENV) && (logLevel >= lvl)
}

func getLoggerLine() string {
	_, fn, line, _ := runtime.Caller(2)
	projFolder := defProjFolder
	envProjFolder := os.Getenv(projFolderKey)

	if len(envProjFolder) != 0 {
		projFolder = "/" + envProjFolder + "/"
	}

	fns := strings.Split(fn, projFolder)
	fn = fns[len(fns)-1]

	return fmt.Sprintf("\t%s:%d  ", fn, line)
}
