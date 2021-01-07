package log

import (
	"log"
	"os"
)

//Logger simple log wrapper.
type Logger struct {
	Warn  *log.Logger
	Info  *log.Logger
	Error *log.Logger
	Debug *log.Logger
}

//MakeLogger create a new logger.
func New(tag string) *Logger {
	return &Logger{
		Warn:  log.New(os.Stderr, "WARNING "+tag+": ", log.Ldate|log.Ltime|log.Lshortfile),
		Info:  log.New(os.Stdout, "INFO "+tag+": ", log.Ldate|log.Ltime|log.Lshortfile),
		Debug: log.New(os.Stderr, "DEBUG "+tag+": ", log.Ldate|log.Ltime|log.Lshortfile),
		Error: log.New(os.Stderr, "ERROR "+tag+": ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}
