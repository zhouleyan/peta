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

package types

import "peta.io/peta/pkg/types/component"

type TypeMeta struct {
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
	//APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

type ObjectMeta struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type Spec struct {
	Components []component.Component `json:"components,omitempty" yaml:"components,omitempty"`
}

type Blueprint struct {
	TypeMeta `json:",inline" yaml:",inline"`
	//ObjectMeta `json:"metadata,omitempty" yaml:"metadata"`

	//Spec Spec `json:"spec,omitempty" yaml:"spec,omitempty"`
}
