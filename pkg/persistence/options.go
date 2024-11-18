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
	"fmt"
	"github.com/spf13/pflag"
	"peta.io/peta/pkg/utils/iputils"
	"strings"
)

const (
	URL = "db-url"
)

type Options struct {
	// `database` determines the name of the database schema to use.
	Database string `json:"database" yaml:"database" mapstructure:"database"`
	// `dialect` is the name of the database system to use.
	Dialect string `json:"dialect" yaml:"dialect" mapstructure:"dialect"`
	// `host` is the host the database is running on.
	Host string `json:"host" yaml:"host" mapstructure:"host"`
	// `port` is the port the database system is running on.
	Port int `json:"port" yaml:"port" mapstructure:"port"`
	// `user` is the database user to use for connecting to the database.
	User string `json:"user" yaml:"user" mapstructure:"user"`
	// `password` is the password of the database user to use for connecting to the database.
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	// `url` is a datasource connecting string. It can be used instead of the rest of the database configuration
	// options. If this `url` is set then it is prioritized, i.e. the rest of the options, if set, have no effect/
	//
	// Schema: `dialect://username:password@host:port/database`
	URL string `json:"url" yaml:"url" mapstructure:"url" jsonSchema:"example=postgres://peta:peta@localhost:5432/peta"`
}

func NewOptions() *Options {
	return &Options{
		Database: "peta",
		Dialect:  "postgres",
		Host:     "localhost",
		Port:     5432,
		User:     "peta",
		Password: "peta",
		URL:      fmt.Sprintf("postgres://peta:peta@localhost:5432/peta"),
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.URL, URL, o.URL, "database connection url")
}

func (o *Options) Merge(fs *pflag.FlagSet, conf *Options) {
	if f := fs.Lookup(URL); f != nil && !f.Changed {
		o.URL = conf.URL
	}
	if conf.Database != "" {
		o.Database = conf.Database
	}
	if conf.Dialect != "" {
		o.Dialect = conf.Dialect
	}
	if conf.Host != "" {
		o.Host = conf.Host
	}
	if conf.Port != 0 {
		o.Port = conf.Port
	}
	if conf.User != "" {
		o.User = conf.User
	}
	if conf.Password != "" {
		o.Password = conf.Password
	}
}

func (o *Options) Validate() []error {
	var errs []error
	if len(strings.TrimSpace(o.URL)) > 0 {
		return errs
	}
	if len(strings.TrimSpace(o.Database)) == 0 {
		errs = append(errs, fmt.Errorf("* database must not be empty"))
	}
	if len(strings.TrimSpace(o.User)) == 0 {
		errs = append(errs, fmt.Errorf("* user must not be empty"))
	}
	if len(strings.TrimSpace(o.Host)) == 0 {
		errs = append(errs, fmt.Errorf("* host must not be empty"))
	}
	if !iputils.IsValidPort(o.Port) {
		errs = append(errs, fmt.Errorf("* port is invalid"))
	}
	if len(strings.TrimSpace(o.Dialect)) == 0 {
		errs = append(errs, fmt.Errorf("* dialect must not be empty"))
	}
	return errs
}
