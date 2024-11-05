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

package sets

import (
	"testing"
)

func TestSets(t *testing.T) {
	testMap := map[string]string{
		"foo_key": "foo_value",
		"bar_key": "bar_value",
	}

	t.Run("set_test", func(t *testing.T) {

		ks := KeySet(testMap)

		ks.Insert("a")
		ks.Insert("a")

		ks = ks.Delete("foo_key")
		for k := range ks {
			if k == "bar_key" {
				t.Logf("%v", ks)
			}

			if k == "foo_key" {
				t.Errorf("%s: set delete failed", t.Name())
			}
		}
	})

	t.Run("set_intersection", func(t *testing.T) {
		s1 := New[string]("foo", "bar")
		s2 := New[string]("bar", "baz")
		s3 := s1.Intersection(s2)
		for k := range s3 {
			if k == "foo" || k == "baz" {
				t.Errorf("%s: intersection failed", t.Name())
				return
			}
			t.Logf("%v", k)
		}
	})
}
