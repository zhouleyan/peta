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

package pathutils

import (
	"peta.io/peta/pkg/utils/splitutils"
	"regexp"
	"strings"
)

func ResolvePath(p string) (name, path string) {
	re := regexp.MustCompile(`\..[a-zA-Z0-9_]+$`)
	if re.MatchString(p) {
		parts := splitutils.SplitPath(p)
		name = splitutils.LastOfSlice(parts)
		re = regexp.MustCompile(`^[^.]*`)
		name = re.FindString(name)
		path = "/" + strings.Join(parts[:len(parts)-1], "/")
	} else {
		path = strings.TrimSuffix(p, "/")
	}
	return name, path
}
