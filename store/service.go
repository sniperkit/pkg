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

package store

import (
	"sync"
	"sync/atomic"

	"github.com/corestoreio/errors"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/config"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/config/cfgmodel"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/store/scope"
)

// Finder depends on the runMode from package scope and finds the active store
// depending on the run mode. The Hash argument will be provided via
// scope.RunMode type or the scope.FromContextRunMode(ctx) function. runMode is
// named in Mage world: MAGE_RUN_CODE and MAGE_RUN_TYPE. The MAGE_RUN_TYPE can
// be either website or store scope and MAGE_RUN_CODE any defined website or
// store code from the database. In our case we must pass an ID and not a code
// string.
type Finder interface {
	// DefaultStoreID returns the default active store ID and its website ID
	// depending on the run mode. Error behaviour is mostly of type NotValid.
	DefaultStoreID(runMode scope.TypeID) (storeID, websiteID int64, err error)
	// StoreIDbyCode returns, depending on the runMode, for a storeCode its
	// active store ID and its website ID. An empty runMode hash falls back to
	// select the default website, with its default group, and the slice of
	// default stores. A not-found error behaviour gets returned if the code
	// cannot be found. If the runMode equals to scope.DefaultTypeID, the
	// returned ID is always 0 and error is nil.
	StoreIDbyCode(runMode scope.TypeID, storeCode string) (storeID, websiteID int64, err error)
}

// Service represents type which handles the underlying storage and takes care
// of the default stores. A Service is bound a specific scope.Scope. Depending
// on the scope it is possible or not to switch stores. A Service contains also
// a config.Getter which gets passed to the scope of a Store(), Group() or
// Website() so that you always have the possibility to access a scoped based
// configuration value. This Service uses three internal maps to cache Websites,
// Groups and Stores.
type Service struct {
	// SingleStoreModeEnabled default value true to enable globally single store
	// mode but might get overwritten via a store scope configuration flag. If
	// this flag is false, single store mode cannot be enabled at all.
	SingleStoreModeEnabled bool

	// BackendSingleStore contains the path to the configuration flag which
	// limits the Service to a single store  Default value: false. Setting this
	// value is optional.
	BackendSingleStore cfgmodel.Bool

	// backend communicates with the database in rw mode and creates
	// new store, group and website pointers. If nil, panics.
	backend *factory
	// defaultStore someone must be always the default guy. Handled via atomic
	// package.
	defaultStoreID int64

	// mu protects the following fields ... maybe use more mutexes
	mu sync.RWMutex
	// in general these caches can be optimized
	websites WebsiteSlice
	groups   GroupSlice
	stores   StoreSlice

	// int64 key identifies a website, group or store
	cacheWebsite     map[int64]Website
	cacheGroup       map[int64]Group
	cacheStore       map[int64]Store
	cacheSingleStore map[scope.TypeID]bool
}

func newService() *Service {
	return &Service{
		SingleStoreModeEnabled: true,
		defaultStoreID:         -1, // means not set, because 0 can be admin store.
		cacheWebsite:           make(map[int64]Website),
		cacheGroup:             make(map[int64]Group),
		cacheStore:             make(map[int64]Store),
		cacheSingleStore:       make(map[scope.TypeID]bool),
	}
}

// NewService creates a new store Service which handles websites, groups and
// stores. You must either provide the functional options or call LoadFromDB()
// to setup the internal cache.
func NewService(cfg config.Getter, opts ...Option) (*Service, error) {
	srv := newService()
	if err := srv.loadFromOptions(cfg, opts...); err != nil {
		return nil, errors.Wrap(err, "[store] NewService.ApplyStorage")
	}
	return srv, nil
}

// MustNewService same as NewService, but panics on error.
func MustNewService(cfg config.Getter, opts ...Option) *Service {
	m, err := NewService(cfg, opts...)
	if err != nil {
		panic(err)
	}
	return m
}

// loadFromOptions main function to set up the internal caches from the factory.
// Does nothing when the options have not been passed.
func (s *Service) loadFromOptions(cfg config.Getter, opts ...Option) error {
	if s == nil {
		s = newService()
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	be, err := newFactory(cfg, opts...)
	if err != nil {
		return errors.Wrap(err, "[store] NewService.NewFactory")
	}

	s.backend = be

	ws, err := s.backend.Websites()
	if err != nil {
		return errors.Wrap(err, "[store] NewService.Websites")
	}
	s.websites = ws
	var wsDefaultCounter = make([]int64, 0, ws.Len())
	ws.Each(func(w Website) {
		s.cacheWebsite[w.Data.WebsiteID] = w
		if w.Data.IsDefault.Valid && w.Data.IsDefault.Bool {
			wsDefaultCounter = append(wsDefaultCounter, w.Data.WebsiteID)
		}
	})
	if ws.Len() > 0 && len(wsDefaultCounter) != 1 {
		return errors.NewNotValidf("[store] NewService: Only one Website can be the default Website. Have: %v. All Website IDs: %v", wsDefaultCounter, ws.IDs())
	}

	gs, err := s.backend.Groups()
	if err != nil {
		return errors.Wrap(err, "[store] NewService.Groups")
	}
	s.groups = gs
	gs.Each(func(g Group) {
		s.cacheGroup[g.Data.GroupID] = g
	})

	ss, err := s.backend.Stores()
	if err != nil {
		return errors.Wrap(err, "[store] NewService.Stores")
	}
	s.stores = ss
	ss.Each(func(str Store) {
		s.cacheStore[str.Data.StoreID] = str
	})
	return nil
}

// IsAllowedStoreID checks if the store ID is allowed within the runMode.
// Returns true on success. An error may occur when the default website and
// store can't be selected. An empty scope.Hash checks the default website with
// its default group and its default stores.
func (s *Service) IsAllowedStoreID(runMode scope.TypeID, storeID int64) (isAllowed bool, storeCode string, _ error) {
	scp, scpID := runMode.Unpack()

	switch scp {
	case scope.Store:
		for _, st := range s.stores {
			if st.IsActive() && st.ID() == storeID {
				return true, st.Code(), nil
			}
		}
		return false, "", nil
	case scope.Group:
		for _, st := range s.stores {
			if st.IsActive() && st.GroupID() == scpID && st.ID() == storeID {
				return true, st.Code(), nil
			}
		}
		return false, "", nil
	case scope.Website:
		for _, st := range s.stores {
			if st.IsActive() && st.WebsiteID() == scpID && st.ID() == storeID {
				return true, st.Code(), nil
			}
		}
		return false, "", nil
	default:
		w, err := s.websites.Default()
		if err != nil {
			return false, "", errors.Wrapf(err, "[store] IsAllowedStoreID.Website.Default Scope %s ID %d", scp, scpID)
		}
		g, err := w.DefaultGroup()
		if err != nil {
			return false, "", errors.Wrapf(err, "[store] IsAllowedStoreID.DefaultGroup Scope %s ID %d", scp, scpID)
		}
		for _, st := range s.stores {
			if st.IsActive() && st.WebsiteID() == w.ID() && st.GroupID() == g.ID() && st.ID() == storeID {
				return true, st.Code(), nil
			}
		}
	}
	return false, "", nil
}

// DefaultStoreID returns the default active store ID depending on the run mode.
// Error behaviour is mostly of type NotValid.
func (s *Service) DefaultStoreID(runMode scope.TypeID) (storeID, websiteID int64, _ error) {
	scp, id := runMode.Unpack()
	switch scp {
	case scope.Store:
		st, err := s.Store(id)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "[store] DefaultStoreID Scope %s ID %d", scp, id)
		}
		if !st.IsActive() {
			return 0, 0, errors.NewNotValidf("[store] DefaultStoreID %s the store ID %d is not active", runMode, st.ID())
		}
		return st.ID(), st.WebsiteID(), nil

	case scope.Group:
		g, err := s.Group(id)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "[store] DefaultStoreID Scope %s ID %d", scp, id)
		}
		st, err := s.Store(g.Data.DefaultStoreID)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "[store] DefaultStoreID Scope %s ID %d", scp, id)
		}
		if !st.IsActive() {
			return 0, 0, errors.NewNotValidf("[store] DefaultStoreID %s the store ID %d is not active", runMode, st.ID())
		}
		return st.ID(), st.WebsiteID(), nil
	}

	var w Website
	if scp == scope.Website {
		var err error
		w, err = s.Website(id)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "[store] DefaultStoreID.Website Scope %s ID %d", scp, id)
		}
	} else {
		var err error
		w, err = s.websites.Default()
		if err != nil {
			return 0, 0, errors.Wrapf(err, "[store] DefaultStoreID.Website.Default Scope %s ID %d", scp, id)
		}
	}
	st, err := w.DefaultStore()
	if err != nil {
		return 0, 0, errors.Wrapf(err, "[store] DefaultStoreID.Website.DefaultStore Scope %s ID %d", scp, id)
	}
	if st.Data == nil || !st.IsActive() {
		return 0, 0, errors.NewNotValidf("[store] DefaultStoreID %s the store ID %d is not active", runMode, st.ID())
	}
	return st.ID(), st.WebsiteID(), nil
}

// StoreIDbyCode returns, depending on the runMode, for a storeCode its active
// store ID and its website ID. An empty runMode hash falls back to select the
// default website, with its default group, and the slice of default stores. A
// not-found error behaviour gets returned if the code cannot be found. If the
// runMode equals to scope.DefaultTypeID, the returned ID is always 0 and error
// is nil. Implements interface Finder.
func (s *Service) StoreIDbyCode(runMode scope.TypeID, storeCode string) (storeID, websiteID int64, err error) {
	if storeCode == "" {
		sID, wID, err := s.DefaultStoreID(0)
		return sID, wID, errors.Wrap(err, "[store] IDbyCode.DefaultStoreID")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	// todo maybe add map cache
	switch runMode.Type() {
	case scope.Store:
		for _, st := range s.stores {
			if st.IsActive() && st.Code() == storeCode {
				return st.ID(), st.WebsiteID(), nil
			}
		}
	case scope.Group:
		for _, st := range s.stores {
			if st.IsActive() && st.GroupID() == runMode.ID() && st.Code() == storeCode {
				return st.ID(), st.WebsiteID(), nil
			}
		}
	case scope.Website:
		for _, st := range s.stores {
			if st.IsActive() && st.WebsiteID() == runMode.ID() && st.Code() == storeCode {
				return st.ID(), st.WebsiteID(), nil
			}
		}
	default:
		w, err := s.websites.Default()
		if err != nil {
			return 0, 0, errors.Wrapf(err, "[store] StoreIDbyCode.Website.Default RunMode %s", runMode)
		}
		g, err := w.DefaultGroup()
		if err != nil {
			return 0, 0, errors.Wrapf(err, "[store] StoreIDbyCode.DefaultGroup RunMode %s", runMode)
		}
		for _, st := range s.stores {
			if st.IsActive() && st.WebsiteID() == w.ID() && st.GroupID() == g.ID() && st.Code() == storeCode {
				return st.ID(), st.WebsiteID(), nil
			}
		}
	}
	return 0, 0, errors.NewNotFoundf("[store] Code %q not found for runMode %s", storeCode, runMode)
}

// AllowedStores creates a new slice containing all active stores depending on
// the current runMode. The returned slice and its pointers are owned by the
// callee.
func (s *Service) AllowedStores(runMode scope.TypeID) (StoreSlice, error) {
	scp, scpID := runMode.Unpack()

	switch scp {
	case scope.Store:
		return s.stores.Filter(func(st Store) bool {
			return st.IsActive()
		}), nil

	case scope.Group:
		return s.stores.Filter(func(st Store) bool {
			return st.IsActive() && st.GroupID() == scpID
		}), nil

	case scope.Website:
		return s.stores.Filter(func(st Store) bool {
			return st.IsActive() && st.WebsiteID() == scpID
		}), nil

	default:
		w, err := s.websites.Default()
		if err != nil {
			return nil, errors.Wrapf(err, "[store] AllowedStores.Website.Default: %s", runMode)
		}
		g, err := w.DefaultGroup()
		if err != nil {
			return nil, errors.Wrapf(err, "[store] AllowedStores.DefaultGroup: %s", runMode)
		}
		return s.stores.Filter(func(st Store) bool {
			return st.IsActive() && st.WebsiteID() == w.ID() && st.GroupID() == g.ID()
		}), nil
	}
}

// HasSingleStore checks if we only have one store view besides the admin store
// view. Mostly used in models to the set store id and in blocks to not display
// the e.g. store switch. Global flag.
func (s *Service) HasSingleStore() bool {
	s.mu.RLock()
	has, ok := s.cacheSingleStore[scope.DefaultTypeID]
	s.mu.RUnlock()
	if ok {
		return has
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	has = s.SingleStoreModeEnabled && s.stores.Len() < 3
	s.cacheSingleStore[scope.DefaultTypeID] = has

	return has
}

// IsSingleStoreMode check if Single-Store mode is enabled in the backend
// configuration and there are less than three Stores. This flag only shows that
// admin does not want to show certain UI components at backend (like store
// switchers etc). Store scope specific flag.
func (s *Service) IsSingleStoreMode(cfg config.Scoped) (bool, error) {

	key := cfg.ScopeID()
	s.mu.RLock()
	has, ok := s.cacheSingleStore[key]
	s.mu.RUnlock()
	if ok {
		return has, nil
	}

	var b = true
	if s.BackendSingleStore.IsSet() {
		var err error
		b, err = s.BackendSingleStore.Get(cfg)
		if err != nil {
			return false, errors.Wrap(err, "[store] Service.IsSingleStoreMode")
		}
	}
	has = s.HasSingleStore() && b
	s.mu.Lock()
	s.cacheSingleStore[key] = has
	s.mu.Unlock()
	return has, nil
}

// Website returns the cached Website from an ID including all of its groups and
// all related stores.
func (s *Service) Website(id int64) (Website, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if cs, ok := s.cacheWebsite[id]; ok {
		return cs, nil
	}
	return Website{}, errors.NewNotFoundf("[store] Cannot find Website ID %d", id)
}

// Websites returns a cached slice containing all Websites with its associated
// groups and stores. You shall not modify the returned slice.
func (s *Service) Websites() WebsiteSlice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.websites
}

// Group returns a cached Group which contains all related stores and its website.
func (s *Service) Group(id int64) (Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if cg, ok := s.cacheGroup[id]; ok {
		return cg, nil
	}
	return Group{}, errors.NewNotFoundf("[store] Cannot find Group ID %d", id)
}

// Groups returns a cached slice containing all  Groups with its associated
// stores and websites. You shall not modify the returned slice.
func (s *Service) Groups() GroupSlice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.groups
}

// Store returns the cached Store view containing its group and its website.
func (s *Service) Store(id int64) (Store, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if cs, ok := s.cacheStore[id]; ok {
		return cs, nil
	}
	return Store{}, errors.NewNotFoundf("[store] Cannot find Store ID %d", id)
}

// Stores returns a cached Store slice containing all related websites and groups.
// You shall not modify the returned slice.
func (s *Service) Stores() StoreSlice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stores
}

// DefaultStoreView returns the overall default store view.
func (s *Service) DefaultStoreView() (Store, error) {
	if s.defaultStoreID >= 0 {
		s.mu.RLock()
		defer s.mu.RUnlock() // bug
		if cs, ok := s.cacheStore[atomic.LoadInt64(&s.defaultStoreID)]; ok {
			return cs, nil
		}
	}

	id, err := s.backend.DefaultStoreID()
	if err != nil {
		return Store{}, errors.Wrap(err, "[store] Service.storage.DefaultStoreView")
	}
	atomic.StoreInt64(&s.defaultStoreID, id)
	return s.Store(id)
}

// LoadFromDB reloads the website, store group and store view data from the database.
// After reloading internal cache will be cleared if there are no errors.
func (s *Service) LoadFromResource(twr TableWebsitesResourcer, tgr TableGroupsResourcer, tsr TableStoresResourcer) error {

	if err := s.backend.LoadFromResource(twr, tgr, tsr); err != nil {
		return errors.Wrap(err, "[store] LoadFromDB.Backend")
	}

	s.ClearCache()

	err := s.loadFromOptions(
		s.backend.rootConfig,
		WithTableWebsites(s.backend.websites...),
		WithTableGroups(s.backend.groups...),
		WithTableStores(s.backend.stores...),
	)
	return errors.Wrap(err, "[store] LoadFromDB.ApplyStorage")
}

// ClearCache resets the internal caches which stores the pointers to Websites,
// Groups or Stores. The ReInit() also uses this method to clear caches before
// the Storage gets reloaded.
func (s *Service) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.cacheWebsite) > 0 {
		for k := range s.cacheWebsite {
			delete(s.cacheWebsite, k)
		}
	}
	if len(s.cacheGroup) > 0 {
		for k := range s.cacheGroup {
			delete(s.cacheGroup, k)
		}
	}
	if len(s.cacheStore) > 0 {
		for k := range s.cacheStore {
			delete(s.cacheStore, k)
		}
	}
	s.cacheSingleStore = make(map[scope.TypeID]bool)
	s.defaultStoreID = -1
	s.websites = nil
	s.groups = nil
	s.stores = nil
}

// IsCacheEmpty returns true if the internal cache is empty.
func (s *Service) IsCacheEmpty() bool {
	return len(s.cacheWebsite) == 0 && len(s.cacheGroup) == 0 && len(s.cacheStore) == 0 &&
		s.defaultStoreID == -1
}
