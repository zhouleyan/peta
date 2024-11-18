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

import (
	"embed"
	"github.com/gobuffalo/pop/v6"
	"strconv"
	"time"
)

//go:embed migrations/*
var migrations embed.FS

var _ Storage = &persister{}

// persister is the Persister interface connecting to the database and capable of doing migrations.
type persister struct {
	Conn *pop.Connection
}

func New(options *Options) (Storage, error) {
	connectionDetails := &pop.ConnectionDetails{
		Pool:            5,
		IdlePool:        0,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 1 * time.Hour,
	}
	if len(options.URL) > 0 {
		connectionDetails.URL = options.URL
	} else {
		connectionDetails.Database = options.Database
		connectionDetails.Dialect = options.Dialect
		connectionDetails.Host = options.Host
		connectionDetails.Port = strconv.Itoa(options.Port)
		connectionDetails.User = options.User
		connectionDetails.Password = options.Password
	}

	conn, err := pop.NewConnection(connectionDetails)
	if err != nil {
		return nil, err
	}

	if err := conn.Open(); err != nil {
		return nil, err
	}

	return &persister{Conn: conn}, nil
}

type Persister interface {
	GetConnection() *pop.Connection
	Transaction(fn func(tx *pop.Connection) error) error
}

type Migrator interface {
	MigrateUp() error
	MigrateDown(int) error
}

type Storage interface {
	Migrator
	Persister
}

func (p *persister) GetConnection() *pop.Connection {
	return p.Conn
}

func (p *persister) Transaction(fn func(tx *pop.Connection) error) error {
	return p.Conn.Transaction(fn)
}

// MigrateUp applies all pending up migrations to the Database
func (p *persister) MigrateUp() error {
	migrationBox, err := pop.NewMigrationBox(migrations, p.Conn)
	if err != nil {
		return err
	}
	err = migrationBox.Up()
	if err != nil {
		return err
	}
	return nil
}

// MigrateDown migrates the Database down by the given number of steps
func (p *persister) MigrateDown(steps int) error {
	migrationBox, err := pop.NewMigrationBox(migrations, p.Conn)
	if err != nil {
		return err
	}
	err = migrationBox.Down(steps)
	if err != nil {
		return err
	}
	return nil
}
