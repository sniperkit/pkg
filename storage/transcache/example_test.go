/*
Sniperkit-Bot
- Status: analyzed
*/

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

package transcache_test

import (
	"encoding/gob"
	"fmt"
	"log"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/storage/transcache"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/storage/transcache/tcbigcache"
)

type P struct {
	X, Y, Z int
	Name    string
}

type Q struct {
	X, Y *int32
	Name string
}

type R struct {
	Name string
	Rune rune
}

func init() {
	gob.Register(P{})
	gob.Register(Q{})
	gob.Register(R{})
}

// This example shows the basic usage of the package: Create the transcache
// processor, set some values, get some, re-prime gob, set values get some.
func ExampleWithPooledEncoder() {

	// Use the gob encoder and prime it with the types.
	tc, err := transcache.NewProcessor(
		// Playing around? Try removing P{}, Q{}, R{} from the next line and see what happens.
		transcache.WithPooledEncoder(transcache.GobCodec{}, P{}, Q{}, R{}),
		tcbigcache.With( /*you can set here bigcache.Config*/ ),
	)
	if err != nil {
		log.Fatalf("NewProcessor error: %+v", err)
	}

	pythagorasKey := []byte(`Pythagoras`)
	if err := tc.Set(pythagorasKey, P{3, 4, 5, "Pythagoras"}); err != nil {
		log.Fatalf("Set error 1: %+v", err)
	}
	treeHouseKey := []byte(`TreeHouse`)
	if err := tc.Set(treeHouseKey, P{1782, 1841, 1922, "Treehouse"}); err != nil {
		log.Fatalf("Set error 2: %+v", err)
	}

	// Get from cache and print the values. Get operations are called more frequently
	// than Set operations so we're simulating that with 5 repetitions.
	for i := 0; i < 5; i++ {
		var q Q
		if err := tc.Get(pythagorasKey, &q); err != nil {
			log.Fatalf("Get error 1: %+v", err)
		}
		fmt.Printf("%q: {%d, %d}\n", q.Name, *q.X, *q.Y)

		if err := tc.Get(treeHouseKey, &q); err != nil {
			log.Fatalf("Get error: %+v", err)
		}
		fmt.Printf("%q: {%d, %d}\n", q.Name, *q.X, *q.Y)
	}

	// We overwrite the previously set values
	if err := tc.Set(pythagorasKey, R{"Pythagoras2", 'P'}); err != nil {
		log.Fatalf("Set error 1: %+v", err)
	}
	if err := tc.Set(treeHouseKey, R{"Treehouse2", 'T'}); err != nil {
		log.Fatalf("Set error 2: %+v", err)
	}

	// Get from cache and print the values. Get operations are called more frequently
	// than Set operations so we're simulating that with 5 repetitions.
	for i := 0; i < 5; i++ {
		var r R
		if err := tc.Get(pythagorasKey, &r); err != nil {
			log.Fatalf("Get error 3: %+v", err)
		}
		fmt.Printf("%q: {%d}\n", r.Name, r.Rune)

		if err := tc.Get(treeHouseKey, &r); err != nil {
			log.Fatalf("Get error: %+v", err)
		}
		fmt.Printf("%q: {%d}\n", r.Name, r.Rune)
	}
	// Output:
	//"Pythagoras": {3, 4}
	//"Treehouse": {1782, 1841}
	//"Pythagoras": {3, 4}
	//"Treehouse": {1782, 1841}
	//"Pythagoras": {3, 4}
	//"Treehouse": {1782, 1841}
	//"Pythagoras": {3, 4}
	//"Treehouse": {1782, 1841}
	//"Pythagoras": {3, 4}
	//"Treehouse": {1782, 1841}
	//"Pythagoras2": {80}
	//"Treehouse2": {84}
	//"Pythagoras2": {80}
	//"Treehouse2": {84}
	//"Pythagoras2": {80}
	//"Treehouse2": {84}
	//"Pythagoras2": {80}
	//"Treehouse2": {84}
	//"Pythagoras2": {80}
	//"Treehouse2": {84}
}
