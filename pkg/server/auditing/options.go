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

package auditing

import "github.com/spf13/pflag"

type Options struct {
	Enable bool `json:"enable" yaml:"enable"`
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Validate() []error {
	errors := make([]error, 0)
	return errors
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enable, "auditing-enabled", o.Enable, "Enable auditing component or not.")
}
