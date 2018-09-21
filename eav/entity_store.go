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

package eav

//
//const maxUint64 = ^uint64(0)
//
//type (
//	entityStoreMap map[uint64]*TableEntityStore
//	once           struct {
//		m    sync.Mutex
//		done uint32
//	}
//)
//
//var (
//	EntityStoreMap   = make(entityStoreMap)
//	entityStoreMutex sync.RWMutex
//	initMapDone      once
//)
//
///*
//   @todo
//   - getIncrementModel e.g. eav/entity_increment_numeric
//   and apply this data to the increment model
//       - getIncrementLastID
//       - getIncrementPrefix
//       - getIncrementPadLength
//       - getIncrementPadChar
//   - feature provide a second increment model: UUID
//
//TODO Increment
//
//If an EAV entity type has an attribute with the code increment_id, and
//no backend model is set for that attribute, the
//Magento\Eav\Model\Entity\Attribute\Backend\Increment model is assigned automatically.
//
//All the increment attribute backend model does is call
//$this->getAttribute()->getEntity()->setNewIncrementId($object)
//in the beforeSave() hook method.
//
//Note: See \Magento\Eav\Model\Entity\Attribute::_getDefaultBackendModel()
//for more automatic backend model assignment rules.
//
//By default, Magento 2 uses only the numeric increment model for customers, orders,
//shipments, invoices and credit memos. The increment models shipped with Magento 2 out
//of the box can be found in the directory app/code/Magento/Eav/Model/Entity/Attribute/Increment/.
//
//*/
//
//func InitEntityStoreMap(dbrSess *dbr.Session) error {
//	if atomic.LoadUint32(&initMapDone.done) == 1 {
//		return ErrStoreMapInitialized
//	}
//
//	initMapDone.m.Lock()
//	defer initMapDone.m.Unlock()
//	if initMapDone.done == 0 {
//		defer atomic.StoreUint32(&initMapDone.done, 1)
//
//		s, err := TableCollection.Table(TableIndexEntityStore)
//		if err != nil {
//			return errors.Wrap(err, "TODO")
//		}
//		var ess TableEntityStoreSlice
//		_, err = dbrSess.
//			Select(s.Columns.FieldNames()...).
//			From(s.Name).
//			LoadStructs(&ess)
//		if err != nil {
//			return errors.Wrap(err, "TODO")
//		}
//
//		for _, es := range ess {
//			EntityStoreMap.Set(es.EntityTypeID, es.StoreID, es)
//		}
//
//		ess = ess[:len(ess)-1] // delete Struct Slice https://code.google.com/p/go-wiki/wiki/SliceTricks
//		return nil
//	}
//	return ErrStoreMapInitialized
//}
//
//func getKey(typeID, storeID int64) uint64 {
//	t, s := uint64(typeID), uint64(storeID)
//	k := t << s
//	if k > maxUint64 {
//		panic("Key size too large") // ?? rethink that
//	}
//	return k
//}
//
//func (m entityStoreMap) Get(typeID, storeID int64) (*TableEntityStore, error) {
//	entityStoreMutex.RLock()
//	defer entityStoreMutex.RUnlock()
//	if es, ok := m[getKey(typeID, storeID)]; ok {
//		return es, nil
//	}
//	return nil, errgo.Newf("Key typeID %d storeID %d not found in entity_type map", typeID, storeID)
//}
//
//func (m entityStoreMap) SetLastIncrementID(typeID, storeID int64, lastIncrementID string) error {
//	if lastIncrementID == "" {
//		return ErrLastIncrementIDEmpty
//	}
//	entityStoreMutex.Lock()
//	defer entityStoreMutex.Unlock()
//	if es, ok := m[getKey(typeID, storeID)]; ok {
//		es.IncrementLastID.String = lastIncrementID
//		// @todo now use a goroutine to permanently save that data
//	}
//	return errgo.Newf("Failed to save! Key typeID %d storeID %d not found in entity_type map", typeID, storeID)
//}
//
//func (m entityStoreMap) Set(typeID, storeID int64, es *TableEntityStore) error {
//	entityStoreMutex.Lock()
//	defer entityStoreMutex.Unlock()
//	*(m[getKey(typeID, storeID)]) = *es // copy pointer
//	return nil
//}
