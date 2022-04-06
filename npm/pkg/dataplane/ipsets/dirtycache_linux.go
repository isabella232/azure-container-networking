package ipsets

import "fmt"

type memberAwareDirtyCache struct {
	// map of prefixed set names to original members
	toAddOrUpdateCache map[string]map[string]struct{}
	// map of prefixed set names to original members
	toDeleteCache map[string]map[string]struct{}
}

func newDirtyCache() dirtyCacheMaintainer {
	return &memberAwareDirtyCache{
		toAddOrUpdateCache: make(map[string]map[string]struct{}),
		toDeleteCache:      make(map[string]map[string]struct{}),
	}
}

func (dc *memberAwareDirtyCache) reset() {
	dc.toAddOrUpdateCache = make(map[string]map[string]struct{})
	dc.toDeleteCache = make(map[string]map[string]struct{})
}

func (dc *memberAwareDirtyCache) create(newSet *IPSet) {
	setName := newSet.Name
	if _, ok := dc.toAddOrUpdateCache[setName]; ok {
		return
	}
	info, ok := dc.toDeleteCache[setName]
	if !ok {
		info = make(map[string]struct{})
	}
	dc.toAddOrUpdateCache[setName] = info
	delete(dc.toDeleteCache, setName)
}

func (dc *memberAwareDirtyCache) update(originalSet *IPSet) {
	putIntoAndRemoveFromOther(originalSet, dc.toAddOrUpdateCache, dc.toDeleteCache)
}

func (dc *memberAwareDirtyCache) delete(originalSet *IPSet) {
	putIntoAndRemoveFromOther(originalSet, dc.toDeleteCache, dc.toAddOrUpdateCache)
}

func putIntoAndRemoveFromOther(originalSet *IPSet, intoCache, fromCache map[string]map[string]struct{}) {
	setName := originalSet.Name
	if _, ok := intoCache[setName]; ok {
		return
	}
	members, ok := fromCache[setName]
	if !ok {
		setType := originalSet.Type
		members = make(map[string]struct{})
		if setType.getSetKind() == HashSet {
			for member := range originalSet.IPPodKey {
				members[member] = struct{}{}
			}
		} else {
			for memberName := range originalSet.MemberIPSets {
				members[memberName] = struct{}{}
			}
		}
	}
	intoCache[setName] = members
	delete(fromCache, setName)
}

func (dc *memberAwareDirtyCache) getSetsToAddOrUpdate() []string {
	setsToAddOrUpdate := make([]string, 0, len(dc.toAddOrUpdateCache))
	for setName := range dc.toAddOrUpdateCache {
		setsToAddOrUpdate = append(setsToAddOrUpdate, setName)
	}
	return setsToAddOrUpdate
}

func (dc *memberAwareDirtyCache) getSetsToDelete() []string {
	setsToDelete := make([]string, 0, len(dc.toDeleteCache))
	for setName := range dc.toDeleteCache {
		setsToDelete = append(setsToDelete, setName)
	}
	return setsToDelete
}

func (dc *memberAwareDirtyCache) numSetsToAddOrUpdate() int {
	return len(dc.toAddOrUpdateCache)
}

func (dc *memberAwareDirtyCache) numSetsToDelete() int {
	return len(dc.toDeleteCache)
}

func (dc *memberAwareDirtyCache) isSetToAddOrUpdate(setName string) bool {
	_, ok := dc.toAddOrUpdateCache[setName]
	return ok
}

func (dc *memberAwareDirtyCache) isSetToDelete(setName string) bool {
	_, ok := dc.toDeleteCache[setName]
	return ok
}

func (dc *memberAwareDirtyCache) printAddOrUpdateCache() string {
	return fmt.Sprintf("%+v", dc.toAddOrUpdateCache)
}

func (dc *memberAwareDirtyCache) printDeleteCache() string {
	return fmt.Sprintf("%+v", dc.toDeleteCache)
}

func (dc *memberAwareDirtyCache) getOriginalMembers(setName string) map[string]struct{} {
	members, ok := dc.toAddOrUpdateCache[setName]
	if !ok {
		members, ok = dc.toDeleteCache[setName]
		if !ok {
			return nil
		}
	}
	return members
}
