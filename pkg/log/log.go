/*
 *  This file is part of PETA.
 *  Copyright (C) 2025 The PETA Authors.
 *  PETA is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  PETA is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with PETA. If not, see <https://www.gnu.org/licenses/>.
 */

package log

import (
	"flag"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

var log Log

var commandLine flag.FlagSet

type Log struct {
	logrus.FieldLogger
	Options

	mu sync.Mutex

	Flush func()
}

type Options struct {
	Verbosity bool

	// TextFormatter
	FullTimestamp   bool   // default true
	ForceColors     bool   // default true
	TimestampFormat string // default time.DateTime

	// JSONFormatter
	LogFile     string // default "peta.log"
	PrettyPrint bool   //default false

	// Log File Rotate
	MaxSize, MaxAge, MaxBackups int

	// Async
	Async bool // default false
}

func Setup() {
	o := &log.Options
	logger := logrus.New()
	if o.Verbosity {
		logger.Level = logrus.TraceLevel
	}
	formatter := &logrus.TextFormatter{
		FullTimestamp:   o.FullTimestamp,
		ForceColors:     o.ForceColors,
		TimestampFormat: o.TimestampFormat,
	}

	logger.SetFormatter(formatter)

	logger.AddHook(&StackHook{
		LogLevels:          []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel},
		Skip:               9,
		MaximumCallerDepth: 25,
	})

	logger.SetOutput(os.Stderr)

	jsonHook := NewJSONHook(o.LogFile, o.Async, o.PrettyPrint, o.MaxSize, o.MaxAge, o.MaxBackups)

	logger.AddHook(jsonHook)

	log.FieldLogger = logger
	log.Flush = jsonHook.Flush
}

func init() {
	commandLine.BoolVar(&log.Verbosity, "v", false, "If true, allows Debug() and Trace() to be logged")
	commandLine.BoolVar(&log.Verbosity, "verbosity", false, "If true, allows Debug() and Trace() to be logged")
	commandLine.BoolVar(&log.FullTimestamp, "full-timestamp", true, "If true, enable logging the full timestamp")
	commandLine.BoolVar(&log.ForceColors, "force-colors", true, "If true, set to true to bypass checking for a TTY before outputting colors")
	commandLine.StringVar(&log.TimestampFormat, "timestamp-format", time.DateTime, "to use for display when a full timestamp is printed")
	commandLine.StringVar(&log.LogFile, "log-file", "peta.log", "If non-empty, write log files in this directory")
	commandLine.BoolVar(&log.PrettyPrint, "pretty-print", false, "If true, will indent all JSON logs")
	commandLine.IntVar(&log.MaxSize, "log-file-size", 10, "the size of the log file before rotating(MB)")
	commandLine.IntVar(&log.MaxAge, "log-age", 28, "the age of the log file before rotating")
	commandLine.IntVar(&log.MaxBackups, "log-backups", 3, "the number of log files to keep")
	commandLine.BoolVar(&log.Async, "async", false, "If true, will print logs asynchronously")
}

func Infoln(args ...interface{}) {
	log.Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warnln(args ...interface{}) {
	log.Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Errorln(args ...interface{}) {
	log.Errorln(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalln(args ...interface{}) {
	log.Fatalln(args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Debugln(args ...interface{}) {
	log.Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Flush() {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.Flush()
}

// InitFlags is for explicitly initializing the flags.
func InitFlags(fs *flag.FlagSet) {
	if fs == nil {
		fs = flag.CommandLine
	}

	commandLine.VisitAll(func(f *flag.Flag) {
		fs.Var(f.Value, f.Name, f.Usage)
	})
}
