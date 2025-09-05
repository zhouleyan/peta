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
	"io"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"peta.io/peta/pkg/utils/queue"
)

func NewJSONHook(logFile string, async, prettyPrint bool, MaxSize, MaxAge, MaxBackups int) *JSONHook {
	l := &lumberjack.Logger{
		Filename: logFile,
		// 10MB
		MaxSize: MaxSize,
		// 28 days
		MaxAge: MaxAge,
		// Maximum number of backup files retained
		MaxBackups: MaxBackups,
		LocalTime:  true,
		Compress:   true,
	}

	jsonFormatter := &JSONFormatter{
		PrettyPrint: prettyPrint,
	}

	var q queue.Queue
	if async {
		q = queue.NewQueue(10, 100)
		q.Run()
	}

	return &JSONHook{
		Writer:    l,
		Formatter: jsonFormatter,
		q:         q,
	}
}

// JSONHook writes logs as JSON and save to files
type JSONHook struct {
	Writer    io.Writer
	Formatter logrus.Formatter
	q         queue.Queue
}

func (j *JSONHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (j *JSONHook) Fire(entry *logrus.Entry) error {
	clone := *entry
	formatted, err := j.Formatter.Format(&clone)
	if err != nil {
		return err
	}

	// writes to files
	if j.q != nil {
		j.q.Push(queue.NewJob(formatted, func(v interface{}) {
			_, err = j.Writer.Write(v.([]byte))
		}))
	} else {
		_, err = j.Writer.Write(formatted)
	}
	return err
}

func (j *JSONHook) Flush() {
	if j.q != nil {
		j.q.Terminate()
	}
}
