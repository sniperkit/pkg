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

package tcbigcache

import (
	"math"
	"testing"

	"github.com/allegro/bigcache"
	"github.com/corestoreio/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/storage/transcache"
)

func TestWithBigCache_Success(t *testing.T) {
	p, err := transcache.NewProcessor(With(), transcache.WithEncoder(transcache.JSONCodec{}))
	if err != nil {
		t.Fatal(err)
	}
	var key = []byte(`key1`)
	if err := p.Set(key, math.Pi); err != nil {
		t.Fatal(err)
	}

	var newVal float64
	if err := p.Get(key, &newVal); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, math.Pi, newVal)

}

func TestWithBigCache_Error(t *testing.T) {
	p, err := transcache.NewProcessor(With(bigcache.Config{
		Shards: 3,
	}))
	assert.Nil(t, p)
	assert.True(t, errors.IsFatal(err), "Error: %+v", err)
}
