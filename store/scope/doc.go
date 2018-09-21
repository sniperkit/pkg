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

// Package scope defines the configuration of scopes default, website, group and
// store.
//
// Outside package scope we refer to type scope.Type as simple just scope and
// scope.TypeID as ScopeID.
//
// The fall back explained from bottom to top:
//     + +-----------+
//     | |  Default  | <---------------+
//     | +-----------+                 |
//     |                               +
//     |                         Falls back to
//     |                               ^
//     |      +------------+ +---------+
//     |      |  Websites  |
//     |      +------------+ <---------+
//     |                               +
//     |                         Falls back to
//     |                               +
//     |            +-----------+      |
//     |            |  Stores   +------+
//     +            +-----------+
//     http://asciiflow.com
//
// A group scope does not make sense in the above schema but is supported by
// other Go types in this package.
package scope
