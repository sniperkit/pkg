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

package backendstore

import (
	"github.com/sniperkit/snk.fork.corestoreio-pkg/config/cfgpath"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/config/element"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/storage/text"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/store/scope"
)

// NewConfigStructure global configuration structure for this package. Used in
// frontend (to display the user all the settings) and in backend (scope checks
// and default values). See the source code of this function for the overall
// available sections, groups and fields.
func NewConfigStructure() (element.Sections, error) {
	return element.MakeSectionsValidated(
		element.Section{
			ID:        cfgpath.MakeRoute("general"),
			Label:     text.Chars(`General`),
			SortOrder: 10,
			Scopes:    scope.PermStore,
			Groups: element.MakeGroups(
				element.Group{
					ID:        cfgpath.MakeRoute("store_information"),
					Label:     text.Chars(`Store Information`),
					SortOrder: 100,
					Scopes:    scope.PermStore,
					Fields: element.MakeFields(
						element.Field{
							// Path: general/store_information/name
							ID:        cfgpath.MakeRoute("name"),
							Label:     text.Chars(`Store Name`),
							Type:      element.TypeText,
							SortOrder: 10,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermStore,
						},

						element.Field{
							// Path: general/store_information/phone
							ID:        cfgpath.MakeRoute("phone"),
							Label:     text.Chars(`Store Phone Number`),
							Type:      element.TypeText,
							SortOrder: 20,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermStore,
						},

						element.Field{
							// Path: general/store_information/hours
							ID:        cfgpath.MakeRoute("hours"),
							Label:     text.Chars(`Store Hours of Operation`),
							Type:      element.TypeText,
							SortOrder: 22,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermStore,
						},

						element.Field{
							// Path: general/store_information/country_id
							ID:         cfgpath.MakeRoute("country_id"),
							Label:      text.Chars(`Country`),
							Type:       element.TypeSelect,
							SortOrder:  25,
							Visible:    element.VisibleYes,
							Scopes:     scope.PermWebsite,
							CanBeEmpty: true,
						},

						element.Field{
							// Path: general/store_information/region_id
							ID:        cfgpath.MakeRoute("region_id"),
							Label:     text.Chars(`Region/State`),
							Type:      element.TypeText,
							SortOrder: 27,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermWebsite,
						},

						element.Field{
							// Path: general/store_information/postcode
							ID:        cfgpath.MakeRoute("postcode"),
							Label:     text.Chars(`ZIP/Postal Code`),
							Type:      element.TypeText,
							SortOrder: 30,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermWebsite,
						},

						element.Field{
							// Path: general/store_information/city
							ID:        cfgpath.MakeRoute("city"),
							Label:     text.Chars(`City`),
							Type:      element.TypeText,
							SortOrder: 45,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermWebsite,
						},

						element.Field{
							// Path: general/store_information/street_line1
							ID:        cfgpath.MakeRoute("street_line1"),
							Label:     text.Chars(`Street Address`),
							Type:      element.TypeText,
							SortOrder: 55,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermWebsite,
						},

						element.Field{
							// Path: general/store_information/street_line2
							ID:        cfgpath.MakeRoute("street_line2"),
							Label:     text.Chars(`Street Address Line 2`),
							Type:      element.TypeText,
							SortOrder: 60,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermWebsite,
						},

						element.Field{
							// Path: general/store_information/merchant_vat_number
							ID:         cfgpath.MakeRoute("merchant_vat_number"),
							Label:      text.Chars(`VAT Number`),
							Type:       element.TypeText,
							SortOrder:  61,
							Visible:    element.VisibleYes,
							Scopes:     scope.PermWebsite,
							CanBeEmpty: true,
						},
					),
				},

				element.Group{
					ID:        cfgpath.MakeRoute("single_store_mode"),
					Label:     text.Chars(`Single-Store Mode`),
					SortOrder: 150,
					Scopes:    scope.PermDefault,
					Fields: element.MakeFields(
						element.Field{
							// Path: general/single_store_mode/enabled
							ID:        cfgpath.MakeRoute("enabled"),
							Label:     text.Chars(`Enable Single-Store Mode`),
							Comment:   text.Chars(`This setting will not be taken into account if the system has more than one store view.`),
							Type:      element.TypeSelect,
							SortOrder: 10,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermDefault,
							Default:   0,
						},
					),
				},
			),
		},
	)
}
