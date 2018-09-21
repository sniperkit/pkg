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

package storemock_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/config/cfgmock"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/config/cfgpath"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/store/scope"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/store/storemock"
)

func TestMustNewStoreAU_ConfigNil(t *testing.T) {
	sAU := storemock.MustNewStoreAU(cfgmock.NewService())
	assert.NotNil(t, sAU)
	assert.NotNil(t, sAU.Config)
	assert.NotNil(t, sAU.Website.Config)

	assert.Exactly(t, int64(5), sAU.Config.storeID)
	assert.Exactly(t, int64(2), sAU.Config.websiteID)

	assert.Exactly(t, int64(0), sAU.Website.Config.storeID)
	assert.Exactly(t, int64(2), sAU.Website.Config.websiteID)

}

func TestMustNewStoreAU_ConfigNonNil(t *testing.T) {
	sAU := storemock.MustNewStoreAU(cfgmock.NewService())
	assert.NotNil(t, sAU)
	assert.NotNil(t, sAU.Config)
	assert.NotNil(t, sAU.Website.Config)
}

func TestMustNewStoreAU_Config(t *testing.T) {
	var configPath = cfgpath.MustMakeByString("aa/bb/cc")

	sm := cfgmock.NewService(cfgmock.PathValue{
		configPath.String():                "DefaultScopeString",
		configPath.BindWebsite(2).String(): "WebsiteScopeString",
		configPath.BindStore(5).String():   "StoreScopeString",
	})
	aust := storemock.MustNewStoreAU(sm)

	haveS, err := aust.Website.Config.String(configPath.Route)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assert.Exactly(t, "WebsiteScopeString", haveS)

	haveS, err = aust.Website.Config.String(configPath.Route, scope.Default)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assert.Exactly(t, "DefaultScopeString", haveS)

	haveS, err = aust.Config.String(configPath.Route)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assert.Exactly(t, "StoreScopeString", haveS)

	haveS, err = aust.Config.String(configPath.Route, scope.Default)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assert.Exactly(t, "DefaultScopeString", haveS)

	assert.Exactly(t, scope.TypeIDs{scope.DefaultTypeID, scope.Website.WithID(2), scope.Store.WithID(5)}, sm.AllInvocations().ScopeIDs())
	assert.Exactly(t, 3, sm.AllInvocations().PathCount())

}
