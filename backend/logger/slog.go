package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type MySLog struct {
	slog *slog.Logger
}

func (s MySLog) Debug(args ...interface{}) {
	if len(args) == 1 {
		s.slog.Debug(args[0].(string))
	} else {
		s.slog.Debug(args[0].(string), args[1:]...)
	}
}

func (s MySLog) Info(args ...interface{}) {
	if len(args) == 1 {
		s.slog.Info(args[0].(string))
	} else {
		s.slog.Info(args[0].(string), args[1:]...)
	}
}

func (s MySLog) Warn(args ...interface{}) {
	if len(args) == 1 {
		s.slog.Warn(args[0].(string))
	} else {
		s.slog.Warn(args[0].(string), args[1:]...)
	}
}

func (s MySLog) Error(args ...interface{}) {
	if len(args) == 1 {
		s.slog.Error(args[0].(string))
	} else {
		s.slog.Error(args[0].(string), args[1:]...)
	}
}

func (s MySLog) Fatal(args ...interface{}) {
	if len(args) == 1 {
		s.slog.Info(args[0].(string))
	} else {
		s.slog.Info(args[0].(string), args[1:]...)
	}
}

func NewMySLog(setLevel string, filePath string) *MySLog {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed To Create Log File: " + filePath)
		panic(err)
	}
	writer := io.MultiWriter(file, os.Stdout)
	level := new(slog.LevelVar)
	switch setLevel {
	case "info":
		level.Set(slog.LevelInfo)
	case "debug":
		level.Set(slog.LevelDebug)
	case "warning":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	}
	handle := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	log := slog.New(handle)
	return &MySLog{slog: log}
}
