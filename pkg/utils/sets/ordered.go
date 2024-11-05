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

package sets

// ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// If future releases of Go add new ordered types,
// this constraint will be modified to include them.
type ordered interface {
	integer | float | ~string
}

// integer is a constraint that permits any integer type.
// If future releases of Go add new predeclared integer types,
// this constraint will be modified to include them.
type integer interface {
	signed | unsigned
}

// float is a constraint that permits any floating-point type.
// If future releases of Go add new predeclared floating-point types,
// this constraint will be modified to include them.
type float interface {
	~float32 | ~float64
}

// signed is a constraint that permits any signed integer type.
// If future releases of Go add new predeclared signed integer types,
// this constraint will be modified to include them.
type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// unsigned is a constraint that permits any unsigned integer type.
// If future releases of Go add new predeclared unsigned integer types,
// this constraint will be modified to include them.
type unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}
