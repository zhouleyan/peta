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

package sliceutils

import (
	"cmp"
	"slices"
)

// LeftDiff (left - right), Elements existing only in left and not in right (duplicates removed)
func LeftDiff[T cmp.Ordered](left, right []T) []T {
	lCopy := slices.Clone(left)
	rCopy := slices.Clone(right)
	slices.Sort(lCopy)
	slices.Sort(rCopy)
	lUnique := slices.Compact(lCopy)
	rUnique := slices.Compact(rCopy)

	var res []T
	for _, v := range lUnique {
		if !slices.Contains(rUnique, v) {
			res = append(res, v)
		}
	}
	return res
}

// RightDiff (right - left), Elements existing only in right and not in left (duplicates removed)
func RightDiff[T cmp.Ordered](left, right []T) []T {
	return LeftDiff(right, left)
}
