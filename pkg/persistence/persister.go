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

package persistence

import "github.com/gobuffalo/pop/v6"

// persister is the Persister interface connecting to the database and capable of doing migrations.
type persister struct {
	Conn *pop.Connection
}

type Persister interface{}

type Migrator interface {
	MigrateUp() error
	MigrateDown(int) error
}

type Storage interface {
	Migrator
	Persister
}
