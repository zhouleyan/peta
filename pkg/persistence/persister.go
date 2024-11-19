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
	"context"
	"embed"
	"github.com/emicklei/go-restful/v3"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"peta.io/peta/pkg/apis/healthz"
	"peta.io/peta/pkg/utils/osutils"
	"peta.io/peta/pkg/utils/resilience"
	"strconv"
	"time"
)

//go:embed migrations/*
var migrations embed.FS

var _ Storage = &persister{}

// defaultInitialPing is the default function that will be called within Storage to make sure
// the database is reachable. It can be injected for test purposes by changing the value
var defaultInitialPing = func(p *persister) error {
	if err := resilience.Retry(5*time.Second, 1*time.Minute, p.Ping); err != nil {
		klog.Errorf("could not ping database: %v", err)
		return errors.WithStack(err)
	}
	return nil
}

// persister is the Persister interface connecting to the database and capable of doing migrations.
type persister struct {
	Conn        *pop.Connection
	initialPing func(r *persister) error
}

func New(ctx context.Context, options *Options) (Storage, error) {
	connectionDetails := &pop.ConnectionDetails{
		Pool:            osutils.MaxParallelism() * 2,
		IdlePool:        osutils.MaxParallelism(),
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

	p := &persister{Conn: conn, initialPing: defaultInitialPing}

	if options.Init {
		klog.Infof("Database %s %s %s connecting...", p.Conn.Dialect.Name(), connectionDetails.Database, connectionDetails.Host)
		if err := p.PingContext(ctx); err != nil {
			return nil, err
		}
	}

	return p, nil
}

type Persister interface {
	GetConnection() *pop.Connection
	Transaction(fn func(tx *pop.Connection) error) error
	Close() error
}

type Migrator interface {
	MigrateUp() error
	MigrateDown(int) error
}

type Storage interface {
	Migrator
	Persister
	healthz.HealthChecker
}

type pinger interface {
	Ping() error
	PingContext(ctx context.Context) error
}

func (p *persister) GetConnection() *pop.Connection {
	return p.Conn
}

func (p *persister) Transaction(fn func(tx *pop.Connection) error) error {
	return p.Conn.Transaction(fn)
}

func (p *persister) Name() string {
	return p.Conn.Dialect.Name() + "-checker"
}

func (p *persister) Check(_ *restful.Request) error {
	return p.Ping()
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

func (p *persister) Ping() error {
	return p.Conn.Store.(pinger).Ping()
}

func (p *persister) Close() error {
	return p.Conn.Close()
}

func (p *persister) PingContext(ctx context.Context) error {
	return p.Conn.Store.(pinger).PingContext(ctx)
}
