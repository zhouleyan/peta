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
	"io"
	"os"
	"peta.io/peta/pkg/utils/queue"
)

func NewJSONHook(logFile string) *JSONHook {
	w, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	jsonFormatter := &JSONFormatter{
		PrettyPrint: false,
	}

	q := queue.NewQueue(10, 100)
	q.Run()

	return &JSONHook{
		Writer:    w,
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
	j.q.Push(queue.NewJob(formatted, func(v interface{}) {
		_, err = j.Writer.Write(v.([]byte))
	}))
	return err
}

func (j *JSONHook) Flush() {
	j.q.Terminate()
}
