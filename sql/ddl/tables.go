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

package ddl

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/corestoreio/errors"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/sql/dml"
)

const (
	PrefixView      = "view_" // If identifier starts with this, it is considered a view.
	MainTable       = "main_table"
	AdditionalTable = "additional_table"
	ScopeTable      = "scope_table"
)

// TableOption applies options and helper functions when creating a new table.
// For example loading column definitions.
type TableOption struct {
	// sortOrder takes care that the options gets applied in the correct order.
	// e.g. column loading can only happen when a table is present.
	sortOrder uint8
	fn        func(*Tables) error
}

// Tables handles all the tables defined for a package. Thread safe.
type Tables struct {
	DB dml.QueryExecPreparer
	// Schema represents the name of the database. Might be empty.
	Schema        string
	previousTable string // the table which has been scanned beforehand
	mu            sync.RWMutex
	// tm a map where key = table name and value the table pointer
	tm map[string]*Table
}

// WithDB sets the DB object to the Tables and all sub Table types to handle the
// database connections. It must be set if other options get used to access the
// DB.
func WithDB(db dml.QueryExecPreparer) TableOption {
	return TableOption{
		sortOrder: 10,
		fn: func(tm *Tables) error {
			tm.DB = db
			tm.mu.Lock()
			defer tm.mu.Unlock()
			for _, t := range tm.tm {
				t.DB = db
			}
			return nil
		},
	}
}

// WithTable inserts a new table to the Tables struct, identified by its index.
// You can optionally specify the columns. What is the reason to use int as the
// table index and not a name? Because table names between M1 and M2 get renamed
// and in a Go SQL code generator script of the CoreStore project, we can
// guarantee that the generated index constant will always stay the same but the
// name of the table differs.
func WithTable(tableName string, cols ...*Column) TableOption {
	return TableOption{
		sortOrder: 1,
		fn: func(tm *Tables) error {
			if err := dml.IsValidIdentifier(tableName); err != nil {
				return errors.WithStack(err)
			}

			if err := tm.Upsert(NewTable(tableName, cols...)); err != nil {
				return errors.Wrap(err, "[ddl] WithNewTable.Tables.Insert")
			}
			return nil
		},
	}
}

// WithCreateTable upserts tables to the current `Tables` object. Either it adds a new
// table/view or overwrites existing entries. Argument `identifierCreateSyntax`
// must be balanced slice where index i is the table/view name and i+1 can be
// either empty or contain the SQL CREATE statement. In case a SQL CREATE
// statement has been supplied, it gets executed otherwise ignored. After table
// initialization the create syntax and the column specifications are getting
// loaded. Write the SQL CREATE statement in upper case.
//		WithCreateTable(
//			"sales_order_history", "CREATE TABLE `sales_order_history` ( ... )", // table gets dropped and recreated
//			"sales_order_stat", "CREATE VIEW `sales_order_stat` AS SELECT ...", // table gets dropped and recreated
//			"sales_order", "", // table/view already exists and gets loaded, NOT dropped.
//		)
func WithCreateTable(ctx context.Context, db dml.QueryExecPreparer, identifierCreateSyntax ...string) TableOption {
	return TableOption{
		sortOrder: 5,
		fn: func(tm *Tables) error {
			tm.mu.Lock()
			defer tm.mu.Unlock()

			lenIDCS := len(identifierCreateSyntax)
			if lenIDCS%2 == 1 {
				return errors.NotValid.Newf("[ddl] WithCreateTable expects a balanced slice, but got %d items.", lenIDCS)
			}

			tvNames := make([]string, 0, lenIDCS/2)
			for i := 0; i < lenIDCS; i = i + 2 {
				// tv = table or view
				tvName := identifierCreateSyntax[i]
				tvCreate := identifierCreateSyntax[i+1]

				if err := dml.IsValidIdentifier(tvName); err != nil {
					return errors.WithStack(err)
				}

				tvNames = append(tvNames, tvName)
				t := NewTable(tvName)
				tm.tm[tvName] = t

				if isCreateStmt(tvName, tvCreate) {
					t.IsView = strings.Contains(tvCreate, " VIEW ") || strings.HasPrefix(tvName, PrefixView)
					t.CreateSyntax = tvCreate
					if err := t.Create(ctx, db); err != nil {
						return errors.WithStack(err)
					}
				}

				if err := t.loadCreateSyntax(ctx, db); err != nil {
					return errors.WithStack(err)
				}

			}
			if db != nil {
				tc, err := LoadColumns(ctx, db, tvNames...)
				if err != nil {
					return errors.WithStack(err)
				}
				for _, n := range tvNames {
					t := tm.tm[n]
					t.Schema = tm.Schema
					t.Columns = tc[n]
					t.update()
				}
			}
			return nil
		},
	}
}

func isCreateStmt(idName, stmt string) bool {
	return stmt != "" && strings.HasPrefix(stmt, "CREATE ") && strings.Contains(stmt, idName)
}

// WithDropTable drops the tables or views listed in argument `tableViewNames`.
// If argument `option` contains the string "DISABLE_FOREIGN_KEY_CHECKS", then foreign keys get disabled
// and at the end re-enabled.
func WithDropTable(ctx context.Context, db dml.QueryExecPreparer, option string, tableViewNames ...string) TableOption {
	return TableOption{
		sortOrder: 2,
		fn: func(tm *Tables) (err error) {
			tm.mu.Lock()
			defer tm.mu.Unlock()

			if option != "" && strings.Contains(strings.ToUpper(option), "DISABLE_FOREIGN_KEY_CHECKS") {
				if _, err = db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
					return errors.WithStack(err)
				}
				defer func() {
					if _, err2 := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=1"); err == nil && err2 != nil {
						err = errors.WithStack(err2)
					}
				}()
			}

			for _, name := range tableViewNames {
				if t, ok := tm.tm[name]; ok {
					if err = t.Drop(ctx, db); err != nil {
						return errors.WithStack(err)
					}
					continue
				}

				if err := dml.IsValidIdentifier(name); err != nil {
					return errors.WithStack(err)
				}
				typ := "TABLE"
				if strings.HasPrefix(name, PrefixView) {
					typ = "VIEW"
				}
				if _, err = db.ExecContext(ctx, "DROP "+typ+" IF EXISTS "+dml.Quoter.Name(name)); err != nil {
					return errors.Wrapf(err, "[ddl] Failed to drop %q", name)
				}
			}

			return nil
		},
	}
}

// WithTableDMLListeners adds event listeners to a table object. It doesn't
// matter if the table has already been set. If the table object gets set later,
// the events will be copied to the new object.
func WithTableDMLListeners(tableName string, events ...*dml.ListenerBucket) TableOption {
	return TableOption{
		sortOrder: 253,
		fn: func(tm *Tables) error {
			tm.mu.Lock()
			defer tm.mu.Unlock()

			t, ok := tm.tm[tableName]
			if !ok {
				return errors.NotFound.Newf("[ddl] Table %q not found", tableName)
			}
			t.Listeners.Merge(events...)
			tm.tm[tableName] = t

			return nil
		},
	}
}

// NewTables creates a new TableService satisfying interface Manager.
func NewTables(opts ...TableOption) (*Tables, error) {
	tm := &Tables{
		tm: make(map[string]*Table),
	}
	if err := tm.Options(opts...); err != nil {
		return nil, errors.WithStack(err)
	}
	return tm, nil
}

// MustNewTables same as NewTableService but panics on error.
func MustNewTables(opts ...TableOption) *Tables {
	ts, err := NewTables(opts...)
	if err != nil {
		panic(err)
	}
	return ts
}

// Options applies options to the Tables service.
func (tm *Tables) Options(opts ...TableOption) error {
	// SliceStable must be stable to maintain the order of all options where
	// sortOrder is zero.
	sort.SliceStable(opts, func(i, j int) bool {
		return opts[i].sortOrder < opts[j].sortOrder
	})

	for _, to := range opts {
		if err := to.fn(tm); err != nil {
			return errors.WithStack(err)
		}
	}
	tm.mu.Lock()
	for _, tbl := range tm.tm {
		if tbl.DB != tm.DB {
			tbl.DB = tm.DB
		}
	}
	tm.mu.Unlock()
	return nil
}

// errTableNotFound provides a custom error behaviour with not capturing the
// stack trace and hence less allocs.
type errTableNotFound string

func (t errTableNotFound) ErrorKind() errors.Kind { return errors.NotFound }
func (t errTableNotFound) Error() string {
	return fmt.Sprintf("[ddl] Table %q not found or not yet added.", string(t))
}

// Table returns the structure from a map m by a giving index i. What is the
// reason to use int as the table index and not a name? Because table names
// between M1 and M2 get renamed and in a Go SQL code generator script of the
// CoreStore project, we can guarantee that the generated index constant will
// always stay the same but the name of the table differs.
func (tm *Tables) Table(name string) (*Table, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if t, ok := tm.tm[name]; ok {
		return t, nil
	}
	return nil, errTableNotFound(name)
}

// MustTable same as Table function but panics when the table cannot be found or
// any other error occurs.
func (tm *Tables) MustTable(name string) *Table {
	t, err := tm.Table(name)
	if err != nil {
		panic(err)
	}
	return t
}

// Tables returns a random list of all available table names. It can append the
// names to the argument slice.
func (tm *Tables) Tables(ret ...string) []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if len(tm.tm) == 0 {
		return ret
	}
	if ret == nil {
		ret = make([]string, 0, len(tm.tm))
	}
	for tn := range tm.tm {
		ret = append(ret, tn)
	}
	return ret
}

// Len returns the number of all tables.
func (tm *Tables) Len() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.tm)
}

// Upsert adds or updates a new table into the internal cache. If a table
// already exists, then the new table gets applied. The ListenerBuckets gets
// merged from the existing table to the new table, they will be appended to the
// new table buckets. Empty columns in the new table gets updated from the
// existing table.
func (tm *Tables) Upsert(tNew *Table) error {

	tm.mu.Lock()
	defer tm.mu.Unlock()

	tOld, ok := tm.tm[tNew.Name]
	if tOld == nil || !ok {
		tm.tm[tNew.Name] = tNew
		return nil
	}

	// for now copy only the events from the existing table
	tNew.Listeners.Merge(&tOld.Listeners)

	if tNew.Schema == "" {
		tNew.Schema = tOld.Schema
	}
	if tNew.Name == "" {
		tNew.Name = tOld.Name
	}
	if len(tNew.Columns) == 0 {
		tNew.Columns = tOld.Columns
	}

	tm.tm[tNew.Name] = tNew.update()
	return nil
}

// DeleteFromCache removes tables by their given indexes. If no index has been passed
// then all entries get removed and the map reinitialized.
func (tm *Tables) DeleteFromCache(tableNames ...string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for _, tn := range tableNames {
		delete(tm.tm, tn)
	}
}

// DeleteAllFromCache clears the internal table cache and resets the map.
func (tm *Tables) DeleteAllFromCache() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// maybe clear each pointer in the Table struct to avoid a memory leak
	tm.tm = make(map[string]*Table)
}

// Validate validates the table names and their column against the current
// database schema. The context is used to maybe cancel the "Load Columns"
// query.
func (tm *Tables) Validate(ctx context.Context) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tblNames := make([]string, 0, len(tm.tm))
	for tn := range tm.tm {
		tblNames = append(tblNames, tn)
	}

	tMap, err := LoadColumns(ctx, tm.DB, tblNames...)
	if err != nil {
		return errors.WithStack(err)
	}
	if have, want := len(tMap), len(tm.tm); have != want {
		return errors.Mismatch.Newf("[ddl] Tables count %d does not match table count %d in database.", want, have)
	}
	dbTableNames := make([]string, 0, len(tMap))
	for tn := range tMap {
		dbTableNames = append(dbTableNames, tn)
	}
	sort.Strings(dbTableNames)

	// TODO compare it that way, that the DB table is the master and Go objects must be updated
	// once they do not match the database version.
	for tn, tbl := range tm.tm {
		dbTblCols, ok := tMap[tn]
		if !ok {
			return errors.NotFound.Newf("[ddl] Table %q not found in database. Available tables: %v", tn, dbTableNames)
		}
		if want, have := len(tbl.Columns), len(dbTblCols); want > have {
			return errors.Mismatch.Newf("[ddl] Table %q has more columns (count %d) than its object (column count %d) in the database.", tn, want, have)
		}
		for idx, c := range tbl.Columns {
			dbCol := dbTblCols[idx]
			if c.Field != dbCol.Field {
				return errors.Mismatch.Newf("[ddl] Table %q with column name %q at index %d does not match database column name %q",
					tn, c.Field, idx, dbCol.Field,
				)
			}
			if c.ColumnType != dbCol.ColumnType {
				return errors.Mismatch.Newf("[ddl] Table %q with Go column name %q does not match MySQL column type. MySQL: %q Go: %q.",
					tn, c.Field, dbCol.ColumnType, c.ColumnType,
				)
			}
			if c.Null != dbCol.Null {
				return errors.Mismatch.Newf("[ddl] Table %q with column name %q does not match MySQL null types. MySQL: %q Go: %q",
					tn, c.Field, dbCol.Null, c.Null,
				)
			}
			// maybe more comparisons
		}
	}

	return nil
}
