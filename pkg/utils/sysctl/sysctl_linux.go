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

package sysctl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Sysctl provides a method to set/get values from /proc/sys - in linux systems
// new interface to set/get values of variables formerly handled by sysctl syscall
// If optional `params` have only one string value - this function will
// set this value into corresponding sysctl variable
func Sysctl(name string, params ...string) (string, error) {
	if len(params) > 1 {
		return "", fmt.Errorf("unexcepted additional parameters")
	} else if len(params) == 1 {
		return setSysctl(name, params[0])
	}
	return getSysctl(name)
}

func getSysctl(name string) (string, error) {
	fullName := filepath.Join("/proc/sys", toNormalName(name))
	data, err := os.ReadFile(fullName)
	if err != nil {
		return "", err
	}

	return string(data[:len(data)-1]), nil
}

func setSysctl(name, value string) (string, error) {
	fullName := filepath.Join("/proc/sys", toNormalName(name))
	if err := os.WriteFile(fullName, []byte(value), 0o644); err != nil {
		return "", err
	}

	return getSysctl(name)
}

// Normalize names by using slash as separator
// Sysctl names can use dots or slashes as separator:
// - if dots are used, dots and slashes are interchanged.
// - if slashes are used, slashes and dots are left intact.
// Separator in use is determined by first occurrence.
func toNormalName(name string) string {
	interchange := false
	for _, c := range name {
		if c == '.' {
			interchange = true
			break
		}
		if c == '/' {
			break
		}
	}

	if interchange {
		r := strings.NewReplacer(".", "/", "/", ".")
		return r.Replace(name)
	}
	return name
}
