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

package runtime

var (
	// ReallyCrash controls the behavior of HandleCrash and defaults to
	// true. It's exposed so components can optionally set false to
	// restore prior behavior. This flag is mostly used for tests to validate
	// crash conditions.
	ReallyCrash = true
)

// Must panics on non-nil errors. Useful to handling programmer level errors.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
