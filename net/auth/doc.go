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

// Package auth provides authentication middleware.
//
// TODO refactor this package to extract business logic from transport medium.
//
// Successful authenticated clients may also retrieve a JSON web token.
// Authentication via basic auth, ACL access control list (for different
// routes), IP based, LDAP, SAML ...
// It can set a  github.com/gorilla/securecookie
//
// ScopedConfig can have an Unauthorized ErrorHandler and next Handler
// When set, all requests with the OPTIONS method will use authentication
// Default: false
// EnableAuthOnOptions bool,
//
// Provide an interface to be used with with multiple authentication sources,
// either social like Google, Facebook, Microsoft Account, LinkedIn, GitHub,
// Twitter, Box, Salesforce, amont others, or enterprise identity systems like
// Windows Azure AD, Google Apps, Active Directory, ADFS or any SAML Identity
// Provider.
//
// TODO: github.com/go-authboss/authboss
// TODO: https://github.com/namsral/multipass for caddy and JWT
// TODO: https://github.com/markbates/goth oauth login for all socials
// TODO: use package authorization to handle roles and rules but do not import
// the package rather define in package auth an interface which requires some
// functions for role and rule handling
// https://blog.komand.com/quick-security-wins-in-golang
// https://github.com/goji/httpauth/blob/master/basic_auth.go
//
// TODO: https://github.com/casbin/casbin Note: The plugins and middleware based
// on Casbin can be found at: https://github.com/casbin Casbin is a powerful and
// efficient open-source access control library for Golang projects. It provides
// support for enforcing authorization based on various access control models.

package auth
