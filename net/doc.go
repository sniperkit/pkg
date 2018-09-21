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

// Package net provides additional network helper functions and in subpackages
// middleware.
//
// Which http router should I use? CoreStore doesn't care because it uses the
// standard library http API. You can choose nearly any router you like.
//
// TODO(CyS) consider the next items:
// - context Package: https://twitter.com/peterbourgon/status/752022730812317696
// - Sessions: https://github.com/alexedwards/scs
// - Form decoding https://github.com/monoculum/formam
// - Kerberos github.com/jcmturner/gokrb5
package net
