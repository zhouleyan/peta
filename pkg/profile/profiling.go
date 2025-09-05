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

package profile

import (
	"os"

	"github.com/pkg/profile"
)

type noop struct{}

// Stop is a noop
func (p noop) Stop() {}

func Profile() interface {
	Stop()
} {
	switch os.Getenv("PROFILING") {
	case "cpu":
		return profile.Start(profile.CPUProfile, profile.NoShutdownHook)
	case "mem":
		return profile.Start(profile.MemProfile, profile.NoShutdownHook)
	case "mutex":
		return profile.Start(profile.MutexProfile, profile.NoShutdownHook)
	case "block":
		return profile.Start(profile.BlockProfile, profile.NoShutdownHook)
	}
	return new(noop)
}

// HelpMessage returns a string explaining how profiling works.
func HelpMessage() string {
	return `- PROFILING: Set "PROFILING=cpu" to enable cpu profiling and "PROFILING=mem" to enable memory profiling.
	It is not possible to do both at the same time. Profiling is disabled per default.

	Example: PROFILING=cpu`
}
