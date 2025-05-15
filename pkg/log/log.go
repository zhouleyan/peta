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

	logger.SetOutput(os.Stdout)

	jsonFormatter := &logrus.JSONFormatter{
		PrettyPrint: false,
	}

	logFile, err := os.OpenFile("peta.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger.AddHook(&JSONHook{
		Writer:    logFile,
		Formatter: jsonFormatter,
	})

	log.FieldLogger = logger
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}
