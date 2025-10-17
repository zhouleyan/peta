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
	"slices"
	"testing"
)

func TestDiff(t *testing.T) {
	cases := []struct {
		name     string
		left     []string
		right    []string
		expected []string
	}{
		{
			name:     "RightDiff",
			right:    []string{"10.1.3.1/24", "10.1.4.1/24", "10.1.3.1/24", "10.1.4.1/24"},
			left:     []string{"10.1.1.1/24", "10.1.2.1/24", "10.1.3.1/24", "10.1.3.1/24"},
			expected: []string{"10.1.4.1/24"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := RightDiff(c.left, c.right)
			if !slices.Equal(res, c.expected) {
				t.Errorf("got %v, want %v", res, c.expected)
			}
		})
	}
}
