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

package options

import (
	"errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"peta.io/peta/pkg/log"
	"peta.io/peta/pkg/utils/pathutils"
	"strings"
	"sync"
)

type Config struct {
	loadOnce sync.Once
	name     string
	path     string
	options  *APIServerOptions
}

func MergeConfig(fs *pflag.FlagSet, o *APIServerOptions) (*APIServerOptions, error) {
	c, err := LoadConfig(o.ConfigFile)
	if err != nil {
		log.Fatalf("failed to load config from disk: %v", err)
	}
	o.Merge(fs, c)
	return o, errors.Join(o.Validate()...)
}

// LoadConfig load config file.
func LoadConfig(path string) (*APIServerOptions, error) {
	name, path := pathutils.ResolvePath(path)
	if name == "" {
		name = defaultConfigName
	}
	// 1.make a config
	viper.SetConfigName(name)
	viper.AddConfigPath(path)

	// load from current working directory, only used for debugging
	viper.AddConfigPath(".")

	// load from env variables
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	c := &Config{
		loadOnce: sync.Once{},
		name:     name,
		path:     path,
		options:  NewAPIServerOptions(),
	}

	// 2.load from disk.
	return c.loadFromDisk()
}

func (c *Config) loadFromDisk() (*APIServerOptions, error) {
	var err error
	c.loadOnce.Do(func() {
		if err = viper.ReadInConfig(); err != nil {
			return
		}
		err = viper.Unmarshal(c.options)
	})
	return c.options, err
}
