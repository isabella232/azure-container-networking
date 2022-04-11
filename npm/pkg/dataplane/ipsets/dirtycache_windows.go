package ipsets

import "fmt"

type dirtyCache struct {
	toAddOrUpdateCache map[string]struct{}
	toDeleteCache      map[string]struct{}
}

func newDirtyCache() dirtyCacheMaintainer {
	dc := &dirtyCache{}
	dc.reset()
	return dc
}

func (dc *dirtyCache) reset() {
	dc.toAddOrUpdateCache = make(map[string]struct{})
	dc.toDeleteCache = make(map[string]struct{})
}

func (dc *dirtyCache) resetAddOrUpdateCache() {
	dc.toAddOrUpdateCache = make(map[string]struct{})
}

func (dc *dirtyCache) create(set *IPSet) {
	putInCacheAndRemoveFromOther(set, dc.toAddOrUpdateCache, dc.toDeleteCache)
}
func (dc *dirtyCache) addMember(set *IPSet, member string) {
	putInCacheAndRemoveFromOther(set, dc.toAddOrUpdateCache, dc.toDeleteCache)
}

func (dc *dirtyCache) deleteMember(set *IPSet, member string) {
	putInCacheAndRemoveFromOther(set, dc.toAddOrUpdateCache, dc.toDeleteCache)
}

func (dc *dirtyCache) destroy(set *IPSet) {
	putInCacheAndRemoveFromOther(set, dc.toDeleteCache, dc.toAddOrUpdateCache)
}

func putInCacheAndRemoveFromOther(set *IPSet, intoCache, fromCache map[string]struct{}) {
	if _, ok := intoCache[set.Name]; ok {
		return
	}
	intoCache[set.Name] = struct{}{}
	delete(fromCache, set.Name)
}

func (dc *dirtyCache) getSetsToAddOrUpdate() map[string]struct{} {
	m := make(map[string]struct{}, 0, len(dc.toAddOrUpdateCache))
	for setName := range dc.toAddOrUpdateCache {
		m[setName] = struct{}{}
	}
	return m
}

func (dc *dirtyCache) getSetsToDelete() map[string]struct{} {
	m := make(map[string]struct{}, 0, len(dc.toDeleteCache))
	for setName := range dc.toDeleteCache {
		m[setName] = struct{}{}
	}
	return m
}

func (dc *dirtyCache) numSetsToAddOrUpdate() int {
	return len(dc.toAddOrUpdateCache)
}

func (dc *dirtyCache) numSetsToDelete() int {
	return len(dc.toDeleteCache)
}

func (dc *dirtyCache) isSetToAddOrUpdate(setName string) bool {
	_, ok := dc.toAddOrUpdateCache[setName]
	return ok
}

func (dc *dirtyCache) isSetToDelete(setName string) bool {
	_, ok := dc.toDeleteCache[setName]
	return ok
}

func (dc *dirtyCache) printAddOrUpdateCache() string {
	return fmt.Sprintf("%+v", dc.toAddOrUpdateCache)
}

func (dc *dirtyCache) printDeleteCache() string {
	return fmt.Sprintf("%+v", dc.toDeleteCache)
}

func (dc *dirtyCache) getMembersToAdd(setName string) map[string]struct{} {
	return nil
}

func (dc *dirtyCache) getMembersToDelete(setName string) map[string]struct{} {
	return nil
}
