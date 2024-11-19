/*
 *  This file is part of PETA.
 *  Copyright (C) 2024 The PETA Authors.
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

package resilience

import (
	"fmt"
	"testing"
	"time"
)

func TestUTCTime(t *testing.T) {
	t.Run("UTC time", func(t *testing.T) {
		utcTime := time.Now().UTC()
		fmt.Printf("当前的UTC时间是：%04d-%02d-%02dT%02d:%02d:%02dZ\n",
			utcTime.Year(), utcTime.Month(), utcTime.Day(),
			utcTime.Hour(), utcTime.Minute(), utcTime.Second())
	})
	t.Run("Local time", func(t *testing.T) {
		utcTime := time.Now()
		fmt.Printf("当前的Local时间是：%04d-%02d-%02dT%02d:%02d:%02dZ\n",
			utcTime.Year(), utcTime.Month(), utcTime.Day(),
			utcTime.Hour(), utcTime.Minute(), utcTime.Second())
	})
}
