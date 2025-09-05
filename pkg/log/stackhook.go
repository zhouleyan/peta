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
	"bytes"
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
)

// StackHook add caller frames
type StackHook struct {
	LogLevels          []logrus.Level
	Skip               int
	MaximumCallerDepth int
}

type Frame uintptr

func (s *StackHook) Levels() []logrus.Level {
	if s.LogLevels == nil || len(s.LogLevels) == 0 {
		return logrus.AllLevels
	}
	return s.LogLevels
}

func (s *StackHook) Fire(entry *logrus.Entry) error {
	buffer := &bytes.Buffer{}

	for i := 0; i < s.MaximumCallerDepth; i++ {
		pc, file, line, ok := runtime.Caller(i + s.Skip)
		if !ok {
			break
		}
		funcName := runtime.FuncForPC(pc).Name()
		buffer.WriteString(fmt.Sprintf("\n%s\n        %s:%d", funcName, file, line))
	}

	entry.Message += buffer.String()
	return nil
}
