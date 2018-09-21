/*
Sniperkit-Bot
- Status: analyzed
*/

// Copyright 2015-2017, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package dmlgen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/sql/ddl"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/storage/null"
)

func TestGetGoPrimitive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		c    ddl.Column
		want string
	}{
		{ddl.Column{Field: `category_id214`, DataType: `bigint`, ColumnType: `bigint unsigned`}, "uint64"},
		{ddl.Column{Field: `category_id224`, DataType: `int`, ColumnType: `bigint`}, "int64"},
		{ddl.Column{Field: `category_id225`, DataType: `int`, ColumnType: `bigint unsigned`}, "uint64"},
		{ddl.Column{Field: `category_id225`, DataType: `int`, Null: "YES", ColumnType: `bigint unsigned`}, "null.Int64"},
		{ddl.Column{Field: `category_id236`, DataType: `int`, Default: null.MakeString(`0`)}, "int64"},
		{ddl.Column{Field: `category_id247`, DataType: `int`, Null: "YES", Default: null.MakeString(`0`)}, "null.Int64"},
		{ddl.Column{Field: `category_id258`, DataType: `int`, Null: "YES", Default: null.MakeString(`0`)}, "null.Int64"},
		{ddl.Column{Field: `is_root_cat269`, DataType: `smallint`, Null: "YES", Default: null.MakeString(`0`)}, "null.Bool"},
		{ddl.Column{Field: `is_root_cat180`, DataType: `smallint`, Null: "YES", Default: null.MakeString(`0`)}, "null.Bool"},
		{ddl.Column{Field: `product_name193`, DataType: `varchar`, Null: "YES", Default: null.MakeString(`0`)}, "null.String"},
		{ddl.Column{Field: `product_name193`, DataType: `varchar`, Null: "YES"}, "null.String"},
		{ddl.Column{Field: `_price_______`, DataType: `decimal`, Null: "YES"}, "dml.Decimal"},
		{ddl.Column{Field: `price`, DataType: `double`, Null: "NO"}, "dml.Decimal"},
		{ddl.Column{Field: `msrp`, DataType: `double`, Null: "NO"}, "dml.Decimal"},
		{ddl.Column{Field: `shipping_adjustment_230`, DataType: `decimal`, Null: "YES"}, "dml.Decimal"},
		{ddl.Column{Field: `shipping_adjustment_241`, DataType: `decimal`, Null: "NO"}, "dml.Decimal"},
		{ddl.Column{Field: `shipping_adjstment_252`, DataType: `decimal`, Null: "YES"}, "dml.Decimal"},
		{ddl.Column{Field: `rate__232`, DataType: `decimal`, Null: "NO"}, "dml.Decimal"},
		{ddl.Column{Field: `rate__233`, DataType: `decimal`, ColumnType: `float unsigned`, Null: "NO"}, "dml.Decimal"},
		{ddl.Column{Field: `grand_absot_233`, DataType: `decimal`, Null: "YES"}, "dml.Decimal"},
		{ddl.Column{Field: `some_currencies_242`, DataType: `decimal`, Default: null.MakeString(`0.0000`)}, "dml.Decimal"},
		{ddl.Column{Field: `weight_252`, DataType: `decimal`, Null: "YES", Default: null.MakeString(`0.0000`)}, "dml.Decimal"},
		{ddl.Column{Field: `weight_263`, DataType: `double`, Default: null.MakeString(`0.0000`)}, "float64"},
		{ddl.Column{Field: `created_at_674`, DataType: `date`, Default: null.MakeString(`0000-00-00`)}, "time.Time"},
		{ddl.Column{Field: `created_at_774`, DataType: `date`, Null: "YES", Default: null.MakeString(`0000-00-00`)}, "null.Time"},
		{ddl.Column{Field: `created_at_874`, DataType: `datetime`, Null: "NO", Default: null.MakeString(`0000-00-00`)}, "time.Time"},
		{ddl.Column{Field: `image001`, DataType: `varbinary`, Null: "NO"}, "[]byte"},
		{ddl.Column{Field: `image002`, DataType: `varbinary`, Null: "YES"}, "[]byte"},
		{ddl.Column{Field: `ok_dude1`, DataType: `bit`, Null: "NO"}, "bool"},
		{ddl.Column{Field: `ok_dude2`, DataType: `bit`, Null: "YES"}, "null.Bool"},
		{ddl.Column{Field: `description_001`, DataType: `varchar`, Null: "YES"}, "null.String"},
		{ddl.Column{Field: `description_002`, DataType: `varchar`, Null: "NO"}, "string"},
		{ddl.Column{Field: `description_003`, DataType: `char`, Null: "YES"}, "null.String"},
		{ddl.Column{Field: `description_004`, DataType: `char`, Null: "NO"}, "string"},
	}
	for _, test := range tests {
		have := toGoTypeNull(&test.c)
		require.Exactly(t, test.want, have, "%#v", test)
	}
}
