// Copyright 2015 CoreStore Authors
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

package eav

import (
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/juju/errgo"
)

type (
	// AttributeBackendModeller defines the attribute backend model @todo
	AttributeBackendModeller interface {
		GetTable() string
		IsStatic() bool
		GetType()
		GetEntityIdField()
		SetValueId(valueId int)
		GetValueId()
		//AfterLoad($object);
		//BeforeSave($object);
		//AfterSave($object);
		//BeforeDelete($object);
		//AfterDelete($object);

		GetEntityValueId(entity *CSEntityType)

		SetEntityValueId(entity *CSEntityType, valueId int)
	}

	// AttributeFrontendModeller defines the attribute frontend model @todo
	AttributeFrontendModeller interface {
		TBD()
	}

	// AttributeSourceModeller defines the source where an attribute can also be stored @todo
	AttributeSourceModeller interface {
		TBD()
	}
)

// GetAttributeSelectSql generates the select query to retrieve full attribute configuration
func GetAttributeSelectSql(dbrSess dbr.SessionRunner, aat EntityTypeAdditionalAttributeTabler, entityTypeID, websiteId int64) (*dbr.SelectBuilder, error) {

	ta, err := GetTableStructure(TableAttribute)
	if err != nil {
		return nil, errgo.Mask(err)
	}
	taa, err := aat.TableAdditionalAttribute()
	if err != nil {
		return nil, errgo.Mask(err)
	}

	selectSql := dbrSess.
		Select(ta.AllColumnAliasQuote("main_table")...).
		From(ta.Name, "main_table").
		Join(
		dbr.JoinTable(taa.Name, "additional_table"),
		taa.ColumnAliasQuote("additional_table"),
		dbr.JoinOn("`additional_table`.`attribute_id` = `main_table`.`attribute_id`"),
		dbr.JoinOn("`main_table`.`entity_type_id` = ?", entityTypeID),
	)

	tew, err := aat.TableEavWebsite()
	if err != nil {
		return nil, errgo.Mask(err)
	}

	if tew != nil {
		const scopeTable = "scope_table"
		l := len(tew.Columns) * 2
		cols := make([]string, l)
		j := 0
		for i := 0; i < l; i = i + 2 {
			cols[i] = scopeTable + "." + tew.Columns[j] // real column name
			cols[i+1] = "scope_" + tew.Columns[j]       // alias column name
			j++
		}

		selectSql.
			LeftJoin(
			dbr.JoinTable(tew.Name, "scope_table"),
			dbr.ColumnAlias(cols...),
			dbr.JoinOn("`scope_table`.`attribute_id` = `main_table`.`attribute_id`"),
			dbr.JoinOn("`scope_table`.`website_id` = ?", websiteId),
		)
	}
	return selectSql, nil
}
