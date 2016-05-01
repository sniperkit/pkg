// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scope

import "github.com/corestoreio/csfw/util/errors"

const maxUint32 = 1<<32 - 1

// Option takes care of the hierarchical level between Website, Group and Store.
// Option can be used as an argument in other functions.
// Instead of the [Website|Group|Store]IDer interface you can
// also provide a [Website|Group|Store]Coder interface.
// Order of scope precedence:
// Website -> Group -> Store. Be sure to set e.g. Website and Group to nil
// if you need initialization for store level.
type Option struct {
	Website WebsiteIDer
	Group   GroupIDer
	Store   StoreIDer
}

// SetByCode depending on the scopeType the code string gets converted into a
// StoreCoder or WebsiteCoder interface and the appropriate struct fields
// get assigned with the *Coder interface. scopeType can only be WebsiteID or
// StoreID because a Group code does not exists.
// Error behaviour: NotSupported
func SetByCode(scp Scope, code string) (o Option, err error) {
	c := MockCode(code)
	// GroupID does not have a scope code
	switch scp {
	case Website:
		o.Website = c
	case Store:
		o.Store = c
	default:
		err = errors.NewNotSupportedf("[scope] Scope: %q Code %q", scp, code)
	}
	return
}

// MustSetByCode same as SetByCode but panics on error. Use only during app
// initialization.
func MustSetByCode(scp Scope, code string) Option {
	so, err := SetByCode(scp, code)
	if err != nil {
		panic(err)
	}
	return so
}

// SetByID depending on the scopeType the scopeID int64 gets converted into a
// [Website|Group|Store]IDer.
// Error behaviour: NotSupported
func SetByID(scp Scope, id int64) (o Option, err error) {
	i := MockID(id)
	// the order of the cases is important
	switch scp {
	case Website:
		o.Website = i
	case Group:
		o.Group = i
	case Store:
		o.Store = i
	default:
		err = errors.NewNotSupportedf("[scope] Scope: %q ID %d", scp, id)
	}
	return
}

// MustSetByID same as SetByID but panics on error. Use only during app
// initialization.
func MustSetByID(scp Scope, id int64) Option {
	so, err := SetByID(scp, id)
	if err != nil {
		panic(err)
	}
	return so
}

// Scope returns the underlying scope ID depending on which struct field is set.
// It maintains the hierarchical order: 1. Website, 2. Group, 3. Store.
// If no field has been set returns DefaultID.
func (o Option) Scope() (s Scope) {
	s = Default
	// the order of the cases is important
	switch {
	case o.Website != nil:
		s = Website
	case o.Group != nil:
		s = Group
	case o.Store != nil:
		s = Store
	}
	return
}

// String is short hand for Option.Scope().String()
func (o Option) String() string {
	return o.Scope().String()
}

// StoreCode extracts the Store code. Checks if the interface StoreCoder
// is available.
func (o Option) StoreCode() (code string) {
	if sc, ok := o.Store.(StoreCoder); ok {
		code = sc.StoreCode()
	}
	return
}

// WebsiteCode extracts the Website code. Checks if the interface WebsiteCoder
// is available.
func (o Option) WebsiteCode() (code string) {
	if wc, ok := o.Website.(WebsiteCoder); ok {
		code = wc.WebsiteCode()
	}
	return
}

// ToUint32 generates a non-unique key from the Option.
// Either the *IDer interfaces gets casted to uint32 or the *Coder
// interface gets hashed via fnv32a.
// If both interfaces (*IDer and *Coder) are nil it returns 0 which is default
// for website, group or store.
// The returned value depends on the hierarchy: 1. website, 2. group, 3. store
// and 4. 0.
func (o Option) ToUint32() uint32 {

	switch {
	case nil != o.Website:
		if wC := o.WebsiteCode(); wC != "" {
			return hashCode(wC)
		}
		if id := o.Website.WebsiteID(); id >= 0 && id < maxUint32 {
			return uint32(id)
		}
	case nil != o.Group:
		if id := o.Group.GroupID(); id >= 0 && id < maxUint32 {
			return uint32(id)
		}
	case nil != o.Store:
		if sC := o.StoreCode(); sC != "" {
			return hashCode(sC)
		}
		if id := o.Store.StoreID(); id >= 0 && id < maxUint32 {
			return uint32(id)
		}
	}

	return 0
}

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fnv implements FNV-1 and FNV-1a, non-cryptographic hash functions
// created by Glenn Fowler, Landon Curt Noll, and Phong Vo.
// See
// https://en.wikipedia.org/wiki/Fowler-Noll-Vo_hash_function.
// fnv32a hash
func hashCode(code string) uint32 {

	data := []byte(code)
	var hash uint32 = 2166136261 // offset
	for _, c := range data {
		hash ^= uint32(c)
		hash *= 16777619 // prime
	}
	return hash
}
