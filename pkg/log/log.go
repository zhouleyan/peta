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
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
	"time"
)

var log Log

type Log struct {
	logrus.FieldLogger
	*Options

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
}

func NewOptions() *Options {
	return &Options{
		Verbosity:       false,
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: time.DateTime,
		LogFile:         "peta.log",
		PrettyPrint:     false,
	}
}

func Setup(o *Options) {
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

	jsonHook := NewJSONHook(o.LogFile, o.PrettyPrint)

	logger.AddHook(jsonHook)

	log.Options = o
	log.FieldLogger = logger
	log.Flush = jsonHook.Flush
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

func Flush() {
	log.Flush()
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Verbosity, "v", o.Verbosity, "If true, allows Debug() and Trace() to be logged")
	fs.BoolVar(&o.Verbosity, "verbosity", o.Verbosity, "If true, allows Debug() and Trace() to be logged")
	fs.BoolVar(&o.FullTimestamp, "full-timestamp", o.FullTimestamp, "If true, enable logging the full timestamp")
	fs.BoolVar(&o.ForceColors, "force-colors", o.ForceColors, "If true, set to true to bypass checking for a TTY before outputting colors")
	fs.StringVar(&o.TimestampFormat, "timestamp-format", o.TimestampFormat, "to use for display when a full timestamp is printed")
	fs.StringVar(&o.LogFile, "log-file", o.LogFile, "If non-empty, write log files in this directory")
	fs.BoolVar(&o.PrettyPrint, "pretty-print", o.PrettyPrint, "If true, will indent all JSON logs")
}
