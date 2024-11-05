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

import "sort"

// Set is a set of the same type elements, implemented via map[comparable]struct{} for minimal memory consumption.
type Set[T comparable] map[T]Empty

// New creates a Set from a list of values.
// NOTE: type param must be explicitly instantiated if given items are empty.
func New[T comparable](items ...T) Set[T] {
	ss := make(Set[T], len(items))
	ss.Insert(items...)
	return ss
}

// KeySet creates a Set from keys of a map[comparable](? extends interface{}).
// If the value passed in is not actually a map, this will panic.
func KeySet[T comparable, V any](theMap map[T]V) Set[T] {
	ret := Set[T]{}
	for keyValue := range theMap {
		ret.Insert(keyValue)
	}
	return ret
}

// Insert adds items to the set.
func (s Set[T]) Insert(items ...T) Set[T] {
	for _, item := range items {
		s[item] = Empty{}
	}
	return s
}

func Insert[T comparable](set Set[T], items ...T) Set[T] {
	return set.Insert(items...)
}

// Delete removes all items from the set.
func (s Set[T]) Delete(items ...T) Set[T] {
	for _, item := range items {
		delete(s, item)
	}
	return s
}

// Clear empties the set.
// It is preferable to replace the set with a newly constructed set.
// but not all callers can do that (when there are other references to the map).
// In some cases the set *won't* be fully cleared, e.g. a Set[float32] containing NaN
// can't be cleared because NaN can't be removed.
// For sets containing items of a type that is reflexive for ==,
// this is optimized to a single call to runtime.mapclear().
func (s Set[T]) Clear() Set[T] {
	for key := range s {
		delete(s, key)
	}
	return s
}

// Has returns true if and only if item is contained in the set.
func (s Set[T]) Has(item T) bool {
	_, contained := s[item]
	return contained
}

// HasAll returns true if and only if all items are contained in the set.
func (s Set[T]) HasAll(items ...T) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// HasAny returns true if any items are contained in the set.
func (s Set[T]) HasAny(items ...T) bool {
	for _, item := range items {
		if s.Has(item) {
			return true
		}
	}
	return false
}

// Clone returns a new set which is a copy of the current set.
func (s Set[T]) Clone() Set[T] {
	result := make(Set[T], len(s))
	for key := range s {
		result.Insert(key)
	}
	return result
}

// Difference returns a set of elements that are not in s2.
func (s Set[T]) Difference(s2 Set[T]) Set[T] {
	result := New[T]()
	for key := range s {
		if !s2.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// SymmetricDifference returns a set of elements which are in either of the sets, but not in their intersection.
func (s Set[T]) SymmetricDifference(s2 Set[T]) Set[T] {
	return s.Difference(s2).Union(s2.Difference(s))
}

// Union returns a new set which includes items in either s1 or s2.
func (s Set[T]) Union(s2 Set[T]) Set[T] {
	result := s.Clone()
	for key := range s2 {
		result.Insert(key)
	}
	return result
}

// Intersection returns a new set which includes the item in BOTH s and s2
func (s Set[T]) Intersection(s2 Set[T]) Set[T] {
	var walk, other Set[T]
	result := New[T]()
	if len(s) < len(s2) {
		walk, other = s, s2
	} else {
		walk, other = s2, s
	}
	for key := range walk {
		if other.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// IsSuperset return true if and only if s is superset of s2.
func (s Set[T]) IsSuperset(s2 Set[T]) bool {
	for item := range s2 {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// Equal returns true if and only if s is equal (as a set) to s2.
// Two sets are equal if their membership is identical.
// (In practice, this means same elements, order doesn't matter)
func (s Set[T]) Equal(s2 Set[T]) bool {
	return len(s) == len(s2) && s.IsSuperset(s2)
}

type sortableSliceOfGeneric[T ordered] []T

func (g sortableSliceOfGeneric[T]) Len() int { return len(g) }

func (g sortableSliceOfGeneric[T]) Less(i, j int) bool { return less[T](g[i], g[j]) }

func (g sortableSliceOfGeneric[T]) Swap(i, j int) { g[i], g[j] = g[j], g[i] }

// List returns the contents as a sorted T slice.
//
// This is a separate function and not a method because not all types supported
// by Generic are ordered and only those can be sorted.
func List[T ordered](s Set[T]) []T {
	res := make(sortableSliceOfGeneric[T], 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	sort.Sort(res)
	return res
}

// UnsortedList returns the slice with contents in random order.
func (s Set[T]) UnsortedList() []T {
	res := make([]T, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	return res
}

// PopAny returns a single element from the set.
func (s Set[T]) PopAny() (T, bool) {
	for key := range s {
		s.Delete(key)
		return key, true
	}
	var zeroValue T
	return zeroValue, false
}

// Len returns the size of the set.
func (s Set[T]) Len() int {
	return len(s)
}

func less[T ordered](lhs, rhs T) bool {
	return lhs < rhs
}
