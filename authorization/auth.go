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

package authorization

import (
	"github.com/sniperkit/snk.fork.corestoreio-pkg/storage/csdb"
)

// TableCollection handles all tables and its columns. init() in generated Go file will set the value.
var TableCollection csdb.TableManager

type UserType uint8

const (
	UserTypeIntegration UserType = iota + 1 // must start with 1 because Magento/Authorization/Model/UserContextInterface.php
	UserTypeAdmin
	UserTypeCustomer
	UserTypeGuest
)
