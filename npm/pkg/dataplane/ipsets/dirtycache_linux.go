package ipsets

import (
	"fmt"

	"github.com/Azure/azure-container-networking/npm/metrics"
	"github.com/Azure/azure-container-networking/npm/util"
	"k8s.io/klog"
)

type dirtyCache struct {
	// all maps have keys of set names and values of members to add/delete
	toCreateCache  map[string]*memberDiff
	toUpdateCache  map[string]*memberDiff
	toDestroyCache map[string]*memberDiff
}

type memberDiff struct {
	membersToAdd    map[string]struct{}
	membersToDelete map[string]struct{}
}

func newMemberDiff() *memberDiff {
	return &memberDiff{
		membersToAdd:    make(map[string]struct{}),
		membersToDelete: make(map[string]struct{}),
	}
}

func newDirtyCache() dirtyCacheInterface {
	dc := &dirtyCache{}
	dc.reset()
	return dc
}

func (dc *dirtyCache) reset() {
	dc.toCreateCache = make(map[string]*memberDiff)
	dc.toUpdateCache = make(map[string]*memberDiff)
	dc.toDestroyCache = make(map[string]*memberDiff)
}

func (dc *dirtyCache) create(set *IPSet) {
	// error checking
	if _, ok := dc.toCreateCache[set.Name]; ok {
		msg := fmt.Sprintf("create: set %s should not already be in the toCreateCache", set.Name)
		klog.Error(msg)
		metrics.SendErrorLogAndMetric(util.IpsmID, msg)
	}
	if _, ok := dc.toUpdateCache[set.Name]; ok {
		msg := fmt.Sprintf("create: set %s should not be in the toUpdateCache", set.Name)
		klog.Error(msg)
		metrics.SendErrorLogAndMetric(util.IpsmID, msg)
	}

	diff, ok := dc.toDestroyCache[set.Name]
	if ok {
		// transfer from toDestroyCache to toUpdateCache and maintain member diff
		dc.toUpdateCache[set.Name] = diff
		delete(dc.toDestroyCache, set.Name)
	} else {
		// put in the toCreateCache and mark all current members as membersToAdd
		var members map[string]struct{}
		if set.Kind == HashSet {
			members = make(map[string]struct{}, len(set.IPPodKey))
			for ip := range set.IPPodKey {
				members[ip] = struct{}{}
			}
		} else {
			members = make(map[string]struct{}, len(set.MemberIPSets))
			for _, memberSet := range set.MemberIPSets {
				members[memberSet.HashedName] = struct{}{}
			}
		}
		dc.toCreateCache[set.Name] = &memberDiff{
			membersToAdd: members,
		}
	}
	fmt.Println("here")
}

func (dc *dirtyCache) addMember(set *IPSet, member string) {
	// error checking
	if dc.isSetToDelete(set.Name) {
		msg := fmt.Sprintf("addMember: set %s should not be in the toDestroyCache", set.Name)
		klog.Error(msg)
		metrics.SendErrorLogAndMetric(util.IpsmID, msg)
	}

	diff := dc.getCreateOrUpdateDiff(set)
	_, ok := diff.membersToDelete[member]
	if ok {
		delete(diff.membersToDelete, member)
	} else {
		diff.membersToAdd[member] = struct{}{}
	}
}

func (dc *dirtyCache) deleteMember(set *IPSet, member string) {
	// error checking
	if dc.isSetToDelete(set.Name) {
		msg := fmt.Sprintf("deleteMember: set %s should not be in the toDestroyCache", set.Name)
		klog.Error(msg)
		metrics.SendErrorLogAndMetric(util.IpsmID, msg)
	}

	diff := dc.getCreateOrUpdateDiff(set)
	_, ok := diff.membersToAdd[member]
	if ok {
		delete(diff.membersToAdd, member)
	} else {
		diff.membersToDelete[member] = struct{}{}
	}
}

func (dc *dirtyCache) getCreateOrUpdateDiff(set *IPSet) *memberDiff {
	diff, ok := dc.toCreateCache[set.Name]
	if !ok {
		diff, ok = dc.toUpdateCache[set.Name]
		if !ok {
			diff = newMemberDiff()
			dc.toUpdateCache[set.Name] = diff
		}
	}
	return diff
}

func (dc *dirtyCache) destroy(set *IPSet) {
	// error checking
	if dc.isSetToDelete(set.Name) {
		msg := fmt.Sprintf("destroy: set %s should not already be in the toDestroyCache", set.Name)
		klog.Error(msg)
		metrics.SendErrorLogAndMetric(util.IpsmID, msg)
	}

	if _, ok := dc.toCreateCache[set.Name]; !ok {
		// modify the diff in the toUpdateCache
		// mark all current members as membersToDelete to accommodate force delete
		if set.Kind == HashSet {
			for ip := range set.IPPodKey {
				dc.deleteMember(set, ip)
			}
		} else {
			for _, memberSet := range set.MemberIPSets {
				dc.deleteMember(set, memberSet.HashedName)
			}
		}
	}
	// put the diff in the toDestroyCache
	diff := dc.getCreateOrUpdateDiff(set)
	dc.toDestroyCache[set.Name] = diff
	delete(dc.toCreateCache, set.Name)
	delete(dc.toUpdateCache, set.Name)
}

func (dc *dirtyCache) getSetsToAddOrUpdate() map[string]struct{} {
	sets := make(map[string]struct{}, len(dc.toCreateCache)+len(dc.toUpdateCache))
	for set := range dc.toCreateCache {
		sets[set] = struct{}{}
	}
	for set := range dc.toUpdateCache {
		sets[set] = struct{}{}
	}
	return sets
}

func (dc *dirtyCache) getSetsToDelete() map[string]struct{} {
	sets := make(map[string]struct{}, len(dc.toDestroyCache))
	for setName := range dc.toDestroyCache {
		sets[setName] = struct{}{}
	}
	return sets
}

func (dc *dirtyCache) numSetsToAddOrUpdate() int {
	return len(dc.toCreateCache) + len(dc.toUpdateCache)
}

func (dc *dirtyCache) numSetsToDelete() int {
	return len(dc.toDestroyCache)
}

func (dc *dirtyCache) isSetToAddOrUpdate(setName string) bool {
	_, ok1 := dc.toCreateCache[setName]
	_, ok2 := dc.toUpdateCache[setName]
	return ok1 || ok2
}

func (dc *dirtyCache) isSetToDelete(setName string) bool {
	_, ok := dc.toDestroyCache[setName]
	return ok
}

func (dc *dirtyCache) printAddOrUpdateCache() string {
	return fmt.Sprintf("[to create: %+v], [to update: %+v]", dc.toCreateCache, dc.toUpdateCache)
}

func (dc *dirtyCache) printDeleteCache() string {
	return fmt.Sprintf("%+v", dc.toDestroyCache)
}

func (dc *dirtyCache) getMembersToAdd(setName string) map[string]struct{} {
	fmt.Println("hey1")
	diff, ok := dc.toCreateCache[setName]
	if ok {
		return diff.membersToAdd
	}
	fmt.Println("hey2")
	diff, ok = dc.toUpdateCache[setName]
	if ok {
		return diff.membersToAdd
	}
	fmt.Println("hey3")
	diff, ok = dc.toDestroyCache[setName]
	if ok {
		return diff.membersToAdd
	}
	return nil
}

func (dc *dirtyCache) getMembersToDelete(setName string) map[string]struct{} {
	diff, ok := dc.toCreateCache[setName]
	if ok {
		return diff.membersToDelete
	}
	diff, ok = dc.toUpdateCache[setName]
	if ok {
		return diff.membersToDelete
	}
	diff, ok = dc.toDestroyCache[setName]
	if ok {
		return diff.membersToDelete
	}
	return nil
}
