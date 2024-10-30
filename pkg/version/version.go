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

package version

import (
	"encoding/json"
	"fmt"
	"runtime"
)

var (
	gitVersion   = "v0.0.0"
	gitCommit    = "unknown"
	gitTreeState = "unknown"
	buildDate    = "unknown"
	gitMajor     = "unknown"
	gitMinor     = "unknown"
)

type Info struct {
	GitVersion   string `json:"gitVersion"`
	GitMajor     string `json:"gitMajor"`
	GitMinor     string `json:"gitMinor"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

func (info Info) String() string {
	jsonString, _ := json.Marshal(info)
	return string(jsonString)
}

// Get returns the overall codebase version. It's for
// detecting what code a binary was built from.
func Get() *Info {
	// These variables typically come from -ldflags settings and
	// in their absence fallback to the default settings
	return &Info{
		GitVersion:   gitVersion,
		GitMajor:     gitMajor,
		GitMinor:     gitMinor,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// SetGitVersion sets the gitVersion, this is mainly for testing use.
func SetGitVersion(v string) {
	gitVersion = v
}
