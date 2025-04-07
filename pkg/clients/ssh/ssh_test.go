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

package ssh

import (
	"fmt"
	"github.com/pkg/errors"
	"testing"
)

func TestSSHRun(t *testing.T) {
	client, err := New(
		"root",
		"10.1.1.21",
		22,
		"123456",
		"",
		"",
		"",
		0,
		true,
		false,
	)
	if err != nil {
		t.Fatal(errors.Wrap(err, "connect error"))
	}

	output, err := client.Run("ls -al")
	if err != nil {
		t.Fatalf("run error: %s", err)
	}
	fmt.Println(string(output))
}

func TestSSH(t *testing.T) {
	t.Run("TestSSHRun", TestSSHRun)
}
