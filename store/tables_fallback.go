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

// +build !mage1,!mage2

// Only include this file IF no specific build tag for mage has been set

// PROTOTYPING for new code generation

package store

import (
	"database/sql"
	"sort"

	"github.com/corestoreio/errors"
	"github.com/corestoreio/log"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/storage/csdb"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/storage/dbr"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/util/null"
)

// TableIndex... is the index to a table. These constants are guaranteed
// to stay the same for all Magento versions. Please access a table via this
// constant instead of the raw table name. TableIndex iotas must start with 0.
const (
	TableIndexStore   = iota // Table: store
	TableIndexGroup          // Table: store_group
	TableIndexWebsite        // Table: store_website
	TableIndexZZZ            // the maximum index, which is not available.
)

// will be initialized in tables_fallback_test.go
//
func init() {
	TableCollection = csdb.MustInitTables(TableCollection,
		csdb.WithTable(TableIndexStore, "store"),
		csdb.WithTable(TableIndexGroup, "store_group"),
		csdb.WithTable(TableIndexWebsite, "store_website"),
	)
	// Don't forget to call TableCollection.ReInit(...) in your code to load the column definitions.
}

// TableStoresResourcer can load a single item or a collection from a backend
// service aka. MySQL. Each implementation must be thread safe.
type TableStoresResourcer interface {
	PrepareSelect() error
	Select(args ...interface{}) (TableStoreSlice, error)
	Insert(TableStoreSlice) (lastInsertID int, err error)
	Update(TableStoreSlice) (affectedRows int, err error)
	Delete(TableStoreSlice) (affectedRows int, err error)
	Close() error
}

//// TableStoresMock used to testing. Satisfies interface TableStoresDBer
//type TableStoresMock struct {
//	SelectFn func() (TableStoreSlice, error)
//	InsertFn func(TableStoreSlice) (lastInsertID int, err error)
//	UpdateFn func(TableStoreSlice) (affectedRows int, err error)
//	DeleteFn func(TableStoreSlice) (affectedRows int, err error)
//}
//
//func (tsm TableStoresMock) Select() (TableStoreSlice, error) {
//	return tsm.SelectFn()
//}
//
//func (tsm TableStoresMock) Insert(ts TableStoreSlice) (lastInsertID int, err error) {
//	return tsm.InsertFn(ts)
//}
//
//func (tsm TableStoresMock) Update(ts TableStoreSlice) (affectedRows int, err error) {
//	return tsm.UpdateFn(ts)
//}
//
//func (tsm TableStoresMock) Delete(ts TableStoreSlice) (affectedRows int, err error) {
//	return tsm.DeleteFn(ts)
//}

// NewTableStoreResource creates a new resource for CRUD operations for table
// `store`. Default logger: Blackhole.
func NewTableStoreResource(db dbr.DBer) *tableStoreResource {
	// accept interfaces but return structs
	t := new(csdb.Table)
	*t = *(TableCollection.MustTable(TableIndexStore)) // shallow copy
	ts := &tableStoreResource{
		Log:   log.BlackHole{},
		DB:    db,
		Table: t,
	}
	return ts
}

// tableStoreResource provides CRUD to the MySQL database
// Generated via tableToStruct.
type tableStoreResource struct {
	Log log.Logger
	DB  dbr.DBer
	// Table contains a shallow copy from the table in TableCollection
	Table *csdb.Table
	// Listeners custom listeners which get applied after the global listeners
	// from TableCollection.
	Listeners dbr.ListenerBucket
	// SliceCap capacity of the TableStoreSlice. Default 10.
	SliceCap int
	stmt     *sql.Stmt
	// Idea: ResurrectPreparer; handles different queries
}

func (ts *tableStoreResource) Close() error {
	if ts.stmt != nil {
		return ts.stmt.Close()
	}
	return nil
}

func (ts *tableStoreResource) PrepareSelect() error {
	sb := &dbr.Select{
		Logger:    ts.Log,
		Preparer:  ts.DB,
		Columns:   dbr.Quoter.ColumnsWithQualifier(csdb.MainTable, "store_id", "code", "website_id", "group_id", "name", "sort_order", "is_active"),
		FromTable: dbr.MakeIdentifier(ts.Table.Name, csdb.MainTable),
	}
	sb.Listeners.Merge(ts.Table.Listeners.Select) // global listeners
	sb.Listeners.Merge(ts.Listeners.Select)       // custom listeners

	var err error
	ts.stmt, err = sb.Prepare()
	return errors.Wrap(err, "[store] ....todo")
}

// Select creates a prepared long lived statement.
// Generated via tableToStruct.
func (ts *tableStoreResource) Select(args ...interface{}) (TableStoreSlice, error) {
	if ts.stmt == nil {
		// run normal query without prepare
	}

	rows, err := ts.stmt.Query()
	if err != nil {
		return nil, errors.Wrap(err, "[store] TableStores.LoadAll.Rows")
	}
	defer rows.Close()

	sCap := 10
	if ts.SliceCap > 0 {
		sCap = ts.SliceCap
	}
	tss := make(TableStoreSlice, 0, sCap)

	for rows.Next() {
		ts := new(TableStore)
		if err := rows.Scan(&ts.StoreID, &ts.Code, &ts.WebsiteID, &ts.GroupID, &ts.Name, &ts.SortOrder, &ts.IsActive); err != nil {
			return nil, errors.Wrap(err, "[store] TableStores.LoadAll.Rows.Scan")
		}
		tss = append(tss, ts)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "[store] TableStores.LoadAll.Rows.Err")
	}

	return tss, nil
}

// SQLInsert inserts all records into the database @todo.
// Generated via tableToStruct.
func (s *tableStoreResource) Insert(tss TableStoreSlice) (int, error) {
	return 0, errors.NewNotSupportedf("[store] Missing implementation")
}

// SQLUpdate updates all record in the database @todo.
// Generated via tableToStruct.
func (s *tableStoreResource) Update(tss TableStoreSlice) (int, error) {
	return 0, errors.NewNotSupportedf("[store] Missing implementation")
}

// SQLDelete deletes all record from the database @todo.
// Generated via tableToStruct.
func (s *tableStoreResource) Delete(tss TableStoreSlice) (int, error) {
	return 0, errors.NewNotSupportedf("[store] Missing implementation")
}

// TableStoreSlice represents a collection of TableStore entities
type TableStoreSlice []*TableStore

// TableStore represents a type for DB table store
// Generated via tableToStruct.
type TableStore struct {
	DB      dbr.DBer            `json:"omit"`
	HookDML *dbr.ListenerBucket `json:"omit"`

	StoreID   int64       `db:"store_id" json:",omitempty"`   // store_id smallint(5) unsigned NOT NULL PRI  auto_increment
	Code      null.String `db:"code" json:",omitempty"`       // code varchar(32) NULL UNI
	WebsiteID int64       `db:"website_id" json:",omitempty"` // website_id smallint(5) unsigned NOT NULL MUL DEFAULT '0'
	GroupID   int64       `db:"group_id" json:",omitempty"`   // group_id smallint(5) unsigned NOT NULL MUL DEFAULT '0'
	Name      string      `db:"name" json:",omitempty"`       // name varchar(255) NOT NULL
	SortOrder int64       `db:"sort_order" json:",omitempty"` // sort_order smallint(5) unsigned NOT NULL  DEFAULT '0'
	IsActive  bool        `db:"is_active" json:",omitempty"`  // is_active smallint(5) unsigned NOT NULL MUL DEFAULT '0'
}

// FindByStoreID searches the primary keys and returns a
// *TableStore if found or nil and false.
// Generated via tableToStruct.
func (s TableStoreSlice) FindByStoreID(
	store_id int64,
) (match *TableStore, found bool) {
	for _, u := range s {
		if u != nil && u.StoreID == store_id {
			match = u
			found = true
			return
		}
	}
	return
}

// FindByCode searches through this unique key and returns
// a *TableStore if found or nil and false.
// Generated via tableToStruct.
func (s TableStoreSlice) FindByCode(code string) (match *TableStore, found bool) {
	for _, u := range s {
		if u != nil && u.Code.String == code {
			match = u
			found = true
			return
		}
	}
	return
}

type sortTableStoreSlice struct {
	TableStoreSlice
	lessFunc func(*TableStore, *TableStore) bool
}

// Less will satisfy the sort.Interface and compares via
// the primary key.
// Generated via tableToStruct.
func (s sortTableStoreSlice) Less(i, j int) bool {
	return s.lessFunc(s.TableStoreSlice[i], s.TableStoreSlice[j])
}

// Sort will sort TableStoreSlice.
// Generated via tableToStruct.
func (s TableStoreSlice) Sort(less func(*TableStore, *TableStore) bool) {
	sort.Sort(sortTableStoreSlice{s, less})
}

// Len returns the length and  will satisfy the sort.Interface.
// Generated via tableToStruct.
func (s TableStoreSlice) Len() int { return len(s) }

// LessPK helper functions for sorting by ascending primary key.
// Can be used as an argument in Sort().
// Generated via tableToStruct.
func (s TableStoreSlice) LessPK(i, j *TableStore) bool {
	return i.StoreID < j.StoreID && 1 == 1
}

// Swap will satisfy the sort.Interface.
// Generated via tableToStruct.
func (s TableStoreSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// FilterThis filters the current slice by predicate f without memory allocation.
// Generated via tableToStruct.
func (s TableStoreSlice) FilterThis(f func(*TableStore) bool) TableStoreSlice {
	b := s[:0]
	for _, x := range s {
		if f(x) {
			b = append(b, x)
		}
	}
	return b
}

// Filter returns a new slice filtered by predicate f.
// Generated via tableToStruct.
func (s TableStoreSlice) Filter(f func(*TableStore) bool) TableStoreSlice {
	sl := make(TableStoreSlice, 0, len(s))
	for _, w := range s {
		if f(w) {
			sl = append(sl, w)
		}
	}
	return sl
}

// FilterNot will return a new TableStoreSlice that does not match
// by calling the function f
// Generated via tableToStruct.
func (s TableStoreSlice) FilterNot(f func(*TableStore) bool) TableStoreSlice {
	sl := make(TableStoreSlice, 0, len(s))
	for _, v := range s {
		if f(v) == false {
			sl = append(sl, v)
		}
	}
	return sl
}

// Each will run function f on all items in TableStoreSlice.
// Generated via tableToStruct.
func (s TableStoreSlice) Each(f func(*TableStore)) TableStoreSlice {
	for i := range s {
		f(s[i])
	}
	return s
}

// Cut will remove items i through j-1.
// Generated via tableToStruct.
func (s *TableStoreSlice) Cut(i, j int) {
	z := *s // copy slice header
	copy(z[i:], z[j:])
	for k, n := len(z)-j+i, len(z); k < n; k++ {
		z[k] = nil // this should avoid the memory leak
	}
	z = z[:len(z)-j+i]
	*s = z
}

// Delete will remove an item from the slice.
// Generated via tableToStruct.
func (s *TableStoreSlice) Delete(i int) {
	z := *s // copy the slice header
	end := len(z) - 1
	s.Swap(i, end)
	copy(z[i:], z[i+1:])
	z[end] = nil // this should avoid the memory leak
	z = z[:end]
	*s = z
}

// Insert will place a new item at position i.
// Generated via tableToStruct.
func (s *TableStoreSlice) Insert(n *TableStore, i int) {
	z := *s // copy the slice header
	z = append(z, &TableStore{})
	copy(z[i+1:], z[i:])
	z[i] = n
	*s = z
}

// Append will add a new item at the end of TableStoreSlice.
// Generated via tableToStruct.
func (s *TableStoreSlice) Append(n ...*TableStore) {
	*s = append(*s, n...)
}

// Prepend will add a new item at the beginning of TableStoreSlice.
// Generated via tableToStruct.
func (s *TableStoreSlice) Prepend(n *TableStore) {
	s.Insert(n, 0)
}

// ExtractStore functions for extracting fields from Store
// slice. Generated via tableToStruct.
type ExtractStore struct {
	StoreID   func() []int64
	Code      func() []string
	WebsiteID func() []int64
	GroupID   func() []int64
	Name      func() []string
	SortOrder func() []int64
	IsActive  func() []bool
}

// Extract extracts from a specified field all values into a slice.
// Generated via tableToStruct.
func (s TableStoreSlice) Extract() ExtractStore {
	return ExtractStore{
		StoreID: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.StoreID)
			}
			return ext
		},
		Code: func() []string {
			ext := make([]string, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.Code.String)
			}
			return ext
		},
		WebsiteID: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.WebsiteID)
			}
			return ext
		},
		GroupID: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.GroupID)
			}
			return ext
		},
		Name: func() []string {
			ext := make([]string, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.Name)
			}
			return ext
		},
		SortOrder: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.SortOrder)
			}
			return ext
		},
		IsActive: func() []bool {
			ext := make([]bool, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.IsActive)
			}
			return ext
		},
	}
}

// TableGroupsResourcer can load a single item or a collection from a backend
// service aka. MySQL. Each implementation must be thread safe.
type TableGroupsResourcer interface {
	Select() (TableGroupSlice, error)
	Insert(TableGroupSlice) (lastInsertID int, err error)
	Update(TableGroupSlice) (affectedRows int, err error)
	Delete(TableGroupSlice) (affectedRows int, err error)
}

// TableGroupSlice represents a collection type for DB table store_group
// Generated via tableToStruct.
type TableGroupSlice []*TableGroup

// TableGroup represents a type for DB table store_group
// Generated via tableToStruct.
type TableGroup struct {
	GroupID        int64  `db:"group_id" json:",omitempty"`         // group_id smallint(5) unsigned NOT NULL PRI  auto_increment
	WebsiteID      int64  `db:"website_id" json:",omitempty"`       // website_id smallint(5) unsigned NOT NULL MUL DEFAULT '0'
	Name           string `db:"name" json:",omitempty"`             // name varchar(255) NOT NULL
	RootCategoryID int64  `db:"root_category_id" json:",omitempty"` // root_category_id int(10) unsigned NOT NULL  DEFAULT '0'
	DefaultStoreID int64  `db:"default_store_id" json:",omitempty"` // default_store_id smallint(5) unsigned NOT NULL MUL DEFAULT '0'
}

//
//// parentSQLSelect fills this slice with data from the database.
//// Generated via tableToStruct.
//func (s *TableGroupSlice) parentSQLSelect(dbrSess dbr.SessionRunner, cbs ...dbr.SelectEvent) (int, error) {
//	return 0, nil // csdb.Load(dbrSess, TableCollection, TableIndexGroup, &(*s), cbs...)
//}
//
//// SQLInsert inserts all records into the database @todo.
//// Generated via tableToStruct.
//func (s *TableGroupSlice) SQLInsert(dbrSess dbr.SessionRunner, cbs ...dbr.InsertEvent) (int, error) {
//	return 0, nil
//}
//
//// SQLUpdate updates all record in the database @todo.
//// Generated via tableToStruct.
//func (s *TableGroupSlice) SQLUpdate(dbrSess dbr.SessionRunner, cbs ...dbr.UpdateEvent) (int, error) {
//	return 0, nil
//}
//
//// SQLDelete deletes all record from the database @todo.
//// Generated via tableToStruct.
//func (s *TableGroupSlice) SQLDelete(dbrSess dbr.SessionRunner, cbs ...dbr.DeleteEvent) (int, error) {
//	return 0, nil
//}

// FindByGroupID searches the primary keys and returns a
// *TableGroup if found or nil and false.
// Generated via tableToStruct.
func (s TableGroupSlice) FindByGroupID(
	group_id int64,
) (match *TableGroup, found bool) {
	for _, u := range s {
		if u != nil && u.GroupID == group_id {
			match = u
			found = true
			return
		}
	}
	return
}

type sortTableGroupSlice struct {
	TableGroupSlice
	lessFunc func(*TableGroup, *TableGroup) bool
}

// Less will satisfy the sort.Interface and compares via
// the primary key.
// Generated via tableToStruct.
func (s sortTableGroupSlice) Less(i, j int) bool {
	return s.lessFunc(s.TableGroupSlice[i], s.TableGroupSlice[j])
}

// Sort will sort TableGroupSlice.
// Generated via tableToStruct.
func (s TableGroupSlice) Sort(less func(*TableGroup, *TableGroup) bool) {
	sort.Sort(sortTableGroupSlice{s, less})
}

// Len returns the length and  will satisfy the sort.Interface.
// Generated via tableToStruct.
func (s TableGroupSlice) Len() int { return len(s) }

// LessPK helper functions for sorting by ascending primary key.
// Can be used as an argument in Sort().
// Generated via tableToStruct.
func (s TableGroupSlice) LessPK(i, j *TableGroup) bool {
	return i.GroupID < j.GroupID && 1 == 1
}

// Swap will satisfy the sort.Interface.
// Generated via tableToStruct.
func (s TableGroupSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// FilterThis filters the current slice by predicate f without memory allocation.
// Generated via tableToStruct.
func (s TableGroupSlice) FilterThis(f func(*TableGroup) bool) TableGroupSlice {
	b := s[:0]
	for _, x := range s {
		if f(x) {
			b = append(b, x)
		}
	}
	return b
}

// Filter returns a new slice filtered by predicate f.
// Generated via tableToStruct.
func (s TableGroupSlice) Filter(f func(*TableGroup) bool) TableGroupSlice {
	sl := make(TableGroupSlice, 0, len(s))
	for _, w := range s {
		if f(w) {
			sl = append(sl, w)
		}
	}
	return sl
}

// FilterNot will return a new TableGroupSlice that does not match
// by calling the function f
// Generated via tableToStruct.
func (s TableGroupSlice) FilterNot(f func(*TableGroup) bool) TableGroupSlice {
	sl := make(TableGroupSlice, 0, len(s))
	for _, v := range s {
		if f(v) == false {
			sl = append(sl, v)
		}
	}
	return sl
}

// Each will run function f on all items in TableGroupSlice.
// Generated via tableToStruct.
func (s TableGroupSlice) Each(f func(*TableGroup)) TableGroupSlice {
	for i := range s {
		f(s[i])
	}
	return s
}

// Cut will remove items i through j-1.
// Generated via tableToStruct.
func (s *TableGroupSlice) Cut(i, j int) {
	z := *s // copy slice header
	copy(z[i:], z[j:])
	for k, n := len(z)-j+i, len(z); k < n; k++ {
		z[k] = nil // this should avoid the memory leak
	}
	z = z[:len(z)-j+i]
	*s = z
}

// Delete will remove an item from the slice.
// Generated via tableToStruct.
func (s *TableGroupSlice) Delete(i int) {
	z := *s // copy the slice header
	end := len(z) - 1
	s.Swap(i, end)
	copy(z[i:], z[i+1:])
	z[end] = nil // this should avoid the memory leak
	z = z[:end]
	*s = z
}

// Insert will place a new item at position i.
// Generated via tableToStruct.
func (s *TableGroupSlice) Insert(n *TableGroup, i int) {
	z := *s // copy the slice header
	z = append(z, &TableGroup{})
	copy(z[i+1:], z[i:])
	z[i] = n
	*s = z
}

// Append will add a new item at the end of TableGroupSlice.
// Generated via tableToStruct.
func (s *TableGroupSlice) Append(n ...*TableGroup) {
	*s = append(*s, n...)
}

// Prepend will add a new item at the beginning of TableGroupSlice.
// Generated via tableToStruct.
func (s *TableGroupSlice) Prepend(n *TableGroup) {
	s.Insert(n, 0)
}

// ExtractGroup functions for extracting fields from Group
// slice. Generated via tableToStruct.
type ExtractGroup struct {
	GroupID        func() []int64
	WebsiteID      func() []int64
	Name           func() []string
	RootCategoryID func() []int64
	DefaultStoreID func() []int64
}

// Extract extracts from a specified field all values into a slice.
// Generated via tableToStruct.
func (s TableGroupSlice) Extract() ExtractGroup {
	return ExtractGroup{
		GroupID: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.GroupID)
			}
			return ext
		},
		WebsiteID: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.WebsiteID)
			}
			return ext
		},
		Name: func() []string {
			ext := make([]string, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.Name)
			}
			return ext
		},
		RootCategoryID: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.RootCategoryID)
			}
			return ext
		},
		DefaultStoreID: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.DefaultStoreID)
			}
			return ext
		},
	}
}

// TableGroupsResourcer can load a single item or a collection from a backend
// service aka. MySQL. Each implementation must be thread safe.
type TableWebsitesResourcer interface {
	Select() (TableWebsiteSlice, error)
	Insert(TableWebsiteSlice) (lastInsertID int, err error)
	Update(TableWebsiteSlice) (affectedRows int, err error)
	Delete(TableWebsiteSlice) (affectedRows int, err error)
}

// TableWebsiteSlice represents a collection type for DB table store_website
// Generated via tableToStruct.
type TableWebsiteSlice []*TableWebsite

// TableWebsite represents a type for DB table store_website
// Generated via tableToStruct.
type TableWebsite struct {
	WebsiteID      int64       `db:"website_id" json:",omitempty"`       // website_id smallint(5) unsigned NOT NULL PRI  auto_increment
	Code           null.String `db:"code" json:",omitempty"`             // code varchar(32) NULL UNI
	Name           null.String `db:"name" json:",omitempty"`             // name varchar(64) NULL
	SortOrder      int64       `db:"sort_order" json:",omitempty"`       // sort_order smallint(5) unsigned NOT NULL MUL DEFAULT '0'
	DefaultGroupID int64       `db:"default_group_id" json:",omitempty"` // default_group_id smallint(5) unsigned NOT NULL MUL DEFAULT '0'
	IsDefault      null.Bool   `db:"is_default" json:",omitempty"`       // is_default smallint(5) unsigned NULL  DEFAULT '0'
}

//// parentSQLSelect fills this slice with data from the database.
//// Generated via tableToStruct.
//func (s *TableWebsiteSlice) parentSQLSelect(dbrSess dbr.SessionRunner, cbs ...dbr.SelectEvent) (int, error) {
//	return 0, nil // csdb.Load(dbrSess, TableCollection, TableIndexWebsite, &(*s), cbs...)
//}
//
//// SQLInsert inserts all records into the database @todo.
//// Generated via tableToStruct.
//func (s *TableWebsiteSlice) SQLInsert(dbrSess dbr.SessionRunner, cbs ...dbr.InsertEvent) (int, error) {
//	return 0, nil
//}
//
//// SQLUpdate updates all record in the database @todo.
//// Generated via tableToStruct.
//func (s *TableWebsiteSlice) SQLUpdate(dbrSess dbr.SessionRunner, cbs ...dbr.UpdateEvent) (int, error) {
//	return 0, nil
//}
//
//// SQLDelete deletes all record from the database @todo.
//// Generated via tableToStruct.
//func (s *TableWebsiteSlice) SQLDelete(dbrSess dbr.SessionRunner, cbs ...dbr.DeleteEvent) (int, error) {
//	return 0, nil
//}

// FindByWebsiteID searches the primary keys and returns a
// *TableWebsite if found or nil and false.
// Generated via tableToStruct.
func (s TableWebsiteSlice) FindByWebsiteID(
	website_id int64,
) (match *TableWebsite, found bool) {
	for _, u := range s {
		if u != nil && u.WebsiteID == website_id {
			match = u
			found = true
			return
		}
	}
	return
}

// FindByCode searches through this unique key and returns
// a *TableWebsite if found or nil and false.
// Generated via tableToStruct.
func (s TableWebsiteSlice) FindByCode(code string) (match *TableWebsite, found bool) {
	for _, u := range s {
		if u != nil && u.Code.String == code {
			match = u
			found = true
			return
		}
	}
	return
}

type sortTableWebsiteSlice struct {
	TableWebsiteSlice
	lessFunc func(*TableWebsite, *TableWebsite) bool
}

// Less will satisfy the sort.Interface and compares via
// the primary key.
// Generated via tableToStruct.
func (s sortTableWebsiteSlice) Less(i, j int) bool {
	return s.lessFunc(s.TableWebsiteSlice[i], s.TableWebsiteSlice[j])
}

// Sort will sort TableWebsiteSlice.
// Generated via tableToStruct.
func (s TableWebsiteSlice) Sort(less func(*TableWebsite, *TableWebsite) bool) {
	sort.Sort(sortTableWebsiteSlice{s, less})
}

// Len returns the length and  will satisfy the sort.Interface.
// Generated via tableToStruct.
func (s TableWebsiteSlice) Len() int { return len(s) }

// LessPK helper functions for sorting by ascending primary key.
// Can be used as an argument in Sort().
// Generated via tableToStruct.
func (s TableWebsiteSlice) LessPK(i, j *TableWebsite) bool {
	return i.WebsiteID < j.WebsiteID && 1 == 1
}

// Swap will satisfy the sort.Interface.
// Generated via tableToStruct.
func (s TableWebsiteSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// FilterThis filters the current slice by predicate f without memory allocation.
// Generated via tableToStruct.
func (s TableWebsiteSlice) FilterThis(f func(*TableWebsite) bool) TableWebsiteSlice {
	b := s[:0]
	for _, x := range s {
		if f(x) {
			b = append(b, x)
		}
	}
	return b
}

// Filter returns a new slice filtered by predicate f.
// Generated via tableToStruct.
func (s TableWebsiteSlice) Filter(f func(*TableWebsite) bool) TableWebsiteSlice {
	sl := make(TableWebsiteSlice, 0, len(s))
	for _, w := range s {
		if f(w) {
			sl = append(sl, w)
		}
	}
	return sl
}

// FilterNot will return a new TableWebsiteSlice that does not match
// by calling the function f
// Generated via tableToStruct.
func (s TableWebsiteSlice) FilterNot(f func(*TableWebsite) bool) TableWebsiteSlice {
	sl := make(TableWebsiteSlice, 0, len(s))
	for _, v := range s {
		if f(v) == false {
			sl = append(sl, v)
		}
	}
	return sl
}

// Each will run function f on all items in TableWebsiteSlice.
// Generated via tableToStruct.
func (s TableWebsiteSlice) Each(f func(*TableWebsite)) TableWebsiteSlice {
	for i := range s {
		f(s[i])
	}
	return s
}

// Cut will remove items i through j-1.
// Generated via tableToStruct.
func (s *TableWebsiteSlice) Cut(i, j int) {
	z := *s // copy slice header
	copy(z[i:], z[j:])
	for k, n := len(z)-j+i, len(z); k < n; k++ {
		z[k] = nil // this should avoid the memory leak
	}
	z = z[:len(z)-j+i]
	*s = z
}

// Delete will remove an item from the slice.
// Generated via tableToStruct.
func (s *TableWebsiteSlice) Delete(i int) {
	z := *s // copy the slice header
	end := len(z) - 1
	s.Swap(i, end)
	copy(z[i:], z[i+1:])
	z[end] = nil // this should avoid the memory leak
	z = z[:end]
	*s = z
}

// Insert will place a new item at position i.
// Generated via tableToStruct.
func (s *TableWebsiteSlice) Insert(n *TableWebsite, i int) {
	z := *s // copy the slice header
	z = append(z, &TableWebsite{})
	copy(z[i+1:], z[i:])
	z[i] = n
	*s = z
}

// Append will add a new item at the end of TableWebsiteSlice.
// Generated via tableToStruct.
func (s *TableWebsiteSlice) Append(n ...*TableWebsite) {
	*s = append(*s, n...)
}

// Prepend will add a new item at the beginning of TableWebsiteSlice.
// Generated via tableToStruct.
func (s *TableWebsiteSlice) Prepend(n *TableWebsite) {
	s.Insert(n, 0)
}

// ExtractWebsite functions for extracting fields from Website
// slice. Generated via tableToStruct.
type ExtractWebsite struct {
	WebsiteID      func() []int64
	Code           func() []string
	Name           func() []string
	SortOrder      func() []int64
	DefaultGroupID func() []int64
	IsDefault      func() []bool
}

// Extract extracts from a specified field all values into a slice.
// Generated via tableToStruct.
func (s TableWebsiteSlice) Extract() ExtractWebsite {
	return ExtractWebsite{
		WebsiteID: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.WebsiteID)
			}
			return ext
		},
		Code: func() []string {
			ext := make([]string, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.Code.String)
			}
			return ext
		},
		Name: func() []string {
			ext := make([]string, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.Name.String)
			}
			return ext
		},
		SortOrder: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.SortOrder)
			}
			return ext
		},
		DefaultGroupID: func() []int64 {
			ext := make([]int64, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.DefaultGroupID)
			}
			return ext
		},
		IsDefault: func() []bool {
			ext := make([]bool, 0, len(s))
			for _, v := range s {
				ext = append(ext, v.IsDefault.Bool)
			}
			return ext
		},
	}
}
