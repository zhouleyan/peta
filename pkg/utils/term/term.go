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

package term

import (
	"fmt"
	"github.com/moby/term"
	"io"
)

// Size returns the current width and height of the user's term. If it isn't a term,
// nil is returned. On error, zero values are returned for width and height.
// Usually w must be the stdout of the process. Stderr won't work.
func Size(w io.Writer) (int, int, error) {
	outFd, isTerminal := term.GetFdInfo(w)
	if !isTerminal {
		return 0, 0, fmt.Errorf("given writer is no term")
	}
	winSize, err := term.GetWinsize(outFd)
	if err != nil {
		return 0, 0, err
	}
	return int(winSize.Width), int(winSize.Height), nil
}
