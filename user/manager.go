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

package user

import (
	"sync"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/config"
)

type Manager struct {
	cr config.Getter

	users UserSlice
	mu    sync.RWMutex
}

// In which case I'd expect the slice of errors to be a 1:1 mapping based on
// index to the passed in IDs (so you could have not found errors or not
// authorized errors etc per user).
// func DeleteUsers(ids []UserID) []error
