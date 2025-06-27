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
	"os"
	"time"
)

var log Log

type Log struct {
	logrus.FieldLogger
	Flush func()
}

func init() {
	logger := logrus.New()
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: time.DateTime,
	}

	logger.SetFormatter(formatter)

	logger.AddHook(&StackHook{
		LogLevels:          []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel},
		Skip:               9,
		MaximumCallerDepth: 25,
	})

	logger.SetOutput(os.Stderr)

	jsonHook := NewJSONHook("peta.log")

	logger.AddHook(jsonHook)

	log.FieldLogger = logger
	log.Flush = jsonHook.Flush
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Flush() {
	log.Flush()
}
