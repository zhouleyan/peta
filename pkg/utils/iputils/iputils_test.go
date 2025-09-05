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

package iputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidIP(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		expected bool
	}{
		"127.0.0.1":   {value: "127.0.0.1", expected: true},
		"127.0.1":     {value: "127.0.1", expected: false},
		"a.baidu.com": {value: "a.baidu.com", expected: false},
		"127.0.0.2":   {value: "127.0.0.2", expected: true},
		"10.1.1.0/24": {value: "10.1.1.0/24", expected: false},
	} {
		t.Run(name, func(t *testing.T) {
			actual := IsValidIP(test.value)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		expected bool
	}{
		"127.0.0.1":      {value: "127.0.0.1", expected: false},
		"localhost":      {value: "localhost", expected: true},
		"pg1-master":     {value: "pg1-master", expected: true},
		"baidu.com":      {value: "baidu.com", expected: true},
		"baidu..com":     {value: "baidu..com", expected: false},
		"baidu.com:8080": {value: "baidu.com:8080", expected: false},
		"a.baidu.com":    {value: "a.baidu.com", expected: true},
		"a.b.baidu.com":  {value: "a.b.baidu.com", expected: true},
	} {
		t.Run(name, func(t *testing.T) {
			actual := IsValidDomain(test.value)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestIPUtils(t *testing.T) {
	t.Run("TestIsValidIP", TestIsValidIP)
	t.Run("TestIsValidDomain", TestIsValidDomain)
}
