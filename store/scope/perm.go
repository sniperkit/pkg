/*
Sniperkit-Bot
- Status: analyzed
*/

// Copyright 2015-present, Cyrill @ Schumacher.fm and the CoreStore contributors
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

import (
	"bytes"
	"strings"

	"github.com/corestoreio/errors"
)

// Perm is a bit set and used for permissions depending on the scope.Type.
// Uint16 should be big enough.
type Perm uint16

// PermStore convenient helper contains all scope permission levels. The
// official core_config_data table and its classes to not support the GroupID
// scope, so that is the reason why PermStore does not have a GroupID.
const PermStore Perm = 1<<Default | 1<<Website | 1<<Store

// PermWebsite convenient helper contains default and website scope permission levels.
const PermWebsite Perm = 1<<Default | 1<<Website

// PermDefault convenient helper contains default scope permission level.
const PermDefault Perm = 1 << Default

// PermStoreReverse convenient helper to enforce hierarchy levels. Only used in
// config.Scoped implementation.
const PermStoreReverse Perm = 1 << Store

// PermWebsiteReverse convenient helper to enforce hierarchy levels. Only used in
// config.Scoped implementation.
const PermWebsiteReverse Perm = 1<<Store | 1<<Website

// MakePerm creates a Perm type based on the input argument which can be either:
// "default","d" or "" for PermDefault, "websites", "website" or "w" for
// PermWebsite OR "stores", "store" or "s" for PermStore. Any other argument
// triggers a NotSupported error.
func MakePerm(name string) (p Perm, err error) {
	switch name {
	case strDefault, "d", "":
		p = PermDefault
	case strWebsites, "w":
		p = PermWebsite
	case strStores, "s":
		p = PermStore
	default:
		err = errors.NotSupported.Newf("[scope] Permission Scope identifier %q not supported. Available: d,w,s", name)
	}
	return
}

// All applies DefaultID, WebsiteID and StoreID scopes
func (bits Perm) All() Perm {
	return bits.Set(Default, Website, Store)
}

// Set takes a variadic amount of Group to set them to Bits
func (bits Perm) Set(scopes ...Type) Perm {
	for _, i := range scopes {
		bits |= 1 << i // (1 << power = 2^power)
	}
	return bits
}

// Top returns the highest stored scope within a Perm. A Perm can consists of 3
// scopes: 1. Default -> 2. Website -> 3. Store Highest scope for a Perm with
// all scopes is: Store.
func (bits Perm) Top() Type {
	switch {
	case bits.Has(Store):
		return Store
	case bits.Has(Website):
		return Website
	}
	return Default
}

// Has checks if a given scope.Type exists within a Perm. Only the first argument
// is supported. Providing no argument assumes the scope.DefaultID.
func (bits Perm) Has(s Type) bool {
	return (bits & Perm(1<<s)) != 0
}

// Human readable representation of the permissions
func (bits Perm) Human(ret ...string) []string {
	if ret == nil {
		ret = make([]string, 0, maxType)
	}
	for i := uint(0); i < uint(maxType); i++ {
		bit := (bits & (1 << i)) != 0
		if bit {
			ret = append(ret, Type(i).String())
		}
	}
	return ret
}

// String readable representation of the permissions
func (bits Perm) String() string {
	switch {
	case bits.Has(Store):
		return strStores
	case bits.Has(Website):
		return strWebsites
	}
	return strDefault
}

// TODO for Go2 implement encoding.TextMarshaler and econding.BinaryMarshaler

var (
	nullByte  = []byte("null")
	quoteByte = []byte(`"`)
)

// MarshalJSON implements json.Marshaler
func (bits Perm) MarshalJSON() ([]byte, error) {
	if bits == 0 {
		return nullByte, nil
	}
	var buf strings.Builder
	buf.WriteByte('"')
	buf.WriteString(bits.String())
	buf.WriteByte('"')
	return []byte(buf.String()), nil
}

// MarshalJSON implements json.Marshaler
func (bits *Perm) UnmarshalJSON(data []byte) error {
	if data == nil {
		*bits = 0
		return nil
	}
	if bytes.HasPrefix(data, quoteByte) {
		data = data[1:]
	}
	if bytes.HasSuffix(data, quoteByte) {
		data = data[:len(data)-1]
	}
	p, err := MakePerm(string(data))
	*bits = p
	return errors.WithStack(err)
}
