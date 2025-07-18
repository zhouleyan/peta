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

package component

type Component struct {
	Name      string   `json:"name" yaml:"name"`
	Type      string   `json:"type" yaml:"type"`
	Enabled   bool     `json:"enabled" yaml:"enabled"`
	Hosts     []Host   `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	Config    Config   `json:"config,omitempty" yaml:"config,omitempty"`
}

type Host struct {
	Name            string            `json:"name,omitempty" yaml:"name,omitempty"`
	Address         string            `json:"address,omitempty" yaml:"address,omitempty"`
	InternalAddress string            `json:"internalAddress,omitempty" yaml:"internalAddress,omitempty"`
	Port            int               `json:"port,omitempty" yaml:"port,omitempty"`
	User            string            `json:"user,omitempty" yaml:"user,omitempty"`
	Password        string            `json:"password,omitempty" yaml:"password,omitempty"`
	PrivateKey      string            `json:"privateKey,omitempty" yaml:"privateKey,omitempty"`
	PrivateKeyPath  string            `json:"privateKeyPath,omitempty" yaml:"privateKeyPath,omitempty"`
	Arch            string            `json:"arch,omitempty" yaml:"arch,omitempty"`
	Timeout         *int64            `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Labels          map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type Config interface {
	GetType() string
}
