package ipsets

import (
	"testing"

	"github.com/Azure/azure-container-networking/npm/util"
	"github.com/stretchr/testify/require"
)

// members are only important for Linux
type dirtyCacheResults struct {
	// map of prefixed name to members to add/delete
	toAddOrUpdate map[string]testDiff
	// map of prefixed name to members to add/delete
	toDestroy map[string]testDiff
}

type testDiff struct {
	toAdd    []string
	toDelete []string
}

const (
	ip     = "1.2.3.4"
	podKey = "pod1"
)

func TestDirtyCacheReset(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	set2 := NewIPSet(NewIPSetMetadata("set2", Namespace))
	set3 := NewIPSet(NewIPSetMetadata("set3", Namespace))
	set4 := NewIPSet(NewIPSetMetadata("set4", Namespace))
	dc := newDirtyCache()
	dc.create(set1)
	dc.addMember(set2, ip)
	dc.deleteMember(set3, "4.4.4.4")
	dc.destroy(set4)
	dc.reset()
	assertDirtyCache(t, dc, &dirtyCacheResults{})
}

func TestDirtyCacheCreate(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	dc := newDirtyCache()
	dc.create(set1)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: map[string]testDiff{
			set1.Name: {},
		},
		toDestroy: nil,
	})
}

// func TestDirtyCacheCreateAfterDelete(t *testing.T) {
// 	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
// 	dc := newDirtyCache()
// 	set1.IPPodKey[ip] = podKey
// 	dc.destroy(set1)
// 	// original members shouldn't get updated
// 	dc.create(set1)
// 	assertDirtyCache(t, dc, &dirtyCacheResults{
// 		toAddOrUpdate: map[string]*memberDiff{
// 			set1.Name: {ip},
// 		},
// 		toDestroy: nil,
// 	})
// }

// func TestDirtyCacheUpdateNew(t *testing.T) {
// 	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
// 	dc := newDirtyCache()
// 	set1.IPPodKey[ip] = podKey
// 	dc.update(set1)
// 	assertDirtyCache(t, dc, &dirtyCacheResults{
// 		toAddOrUpdate: map[string][]string{
// 			set1.Name: {ip},
// 		},
// 		toDestroy: nil,
// 	})
// }

// func TestDirtyCacheUpdateOld(t *testing.T) {
// 	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
// 	dc := newDirtyCache()
// 	dc.create(set1)
// 	// original members shouldn't get updated
// 	set1.IPPodKey[ip] = podKey
// 	dc.update(set1)
// 	assertDirtyCache(t, dc, &dirtyCacheResults{
// 		toAddOrUpdate: map[string][]string{
// 			set1.Name: {},
// 		},
// 		toDestroy: nil,
// 	})
// }

// func TestDirtyCacheUpdateTwice(t *testing.T) {
// 	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
// 	dc := newDirtyCache()
// 	dc.update(set1)
// 	// original members shouldn't get updated
// 	set1.IPPodKey[ip] = podKey
// 	dc.update(set1)
// 	assertDirtyCache(t, dc, &dirtyCacheResults{
// 		toAddOrUpdate: map[string][]string{
// 			set1.Name: {},
// 		},
// 		toDestroy: nil,
// 	})
// }

// func TestDirtyCacheUpdateAfterDelete(t *testing.T) {
// 	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
// 	dc := newDirtyCache()
// 	dc.destroy(set1)
// 	// original members shouldn't get updated
// 	set1.IPPodKey[ip] = podKey
// 	dc.update(set1)
// 	assertDirtyCache(t, dc, &dirtyCacheResults{
// 		toAddOrUpdate: map[string][]string{
// 			set1.Name: {},
// 		},
// 		toDestroy: nil,
// 	})
// }

// func TestDirtyCacheDelete(t *testing.T) {
// 	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
// 	dc := newDirtyCache()
// 	set1.IPPodKey[ip] = podKey
// 	dc.destroy(set1)
// 	assertDirtyCache(t, dc, &dirtyCacheResults{
// 		toAddOrUpdate: nil,
// 		toDestroy: map[string][]string{
// 			set1.Name: {ip},
// 		},
// 	})
// }

// func TestDirtyCacheDeleteOld(t *testing.T) {
// 	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
// 	dc := newDirtyCache()
// 	set1.IPPodKey[ip] = podKey
// 	dc.update(set1)
// 	// original members shouldn't get updated
// 	delete(set1.IPPodKey, ip)
// 	dc.destroy(set1)
// 	assertDirtyCache(t, dc, &dirtyCacheResults{
// 		toAddOrUpdate: nil,
// 		toDestroy: map[string][]string{
// 			set1.Name: {ip},
// 		},
// 	})
// }

// func TestDirtyCacheDeleteTwice(t *testing.T) {
// 	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
// 	dc := newDirtyCache()
// 	set1.IPPodKey[ip] = podKey
// 	dc.update(set1)
// 	// original members shouldn't get updated
// 	delete(set1.IPPodKey, ip)
// 	dc.destroy(set1)
// 	assertDirtyCache(t, dc, &dirtyCacheResults{
// 		toAddOrUpdate: nil,
// 		toDestroy: map[string][]string{
// 			set1.Name: {ip},
// 		},
// 	})
// }

func assertDirtyCache(t *testing.T, dc dirtyCacheInterface, expected *dirtyCacheResults) {
	require.Equal(t, len(expected.toAddOrUpdate), dc.numSetsToAddOrUpdate(), "unexpected number of sets to add or update")
	require.Equal(t, len(expected.toDestroy), dc.numSetsToDelete(), "unexpected number of sets to delete")
	for setName, diff := range expected.toAddOrUpdate {
		require.True(t, dc.isSetToAddOrUpdate(setName), "set %s should be added/updated", setName)
		require.False(t, dc.isSetToDelete(setName), "set %s should not be deleted", setName)
		assertDiff(t, dc, setName, diff)
	}
	for setName, diff := range expected.toDestroy {
		require.True(t, dc.isSetToDelete(setName), "set %s should be deleted", setName)
		require.False(t, dc.isSetToAddOrUpdate(setName), "set %s should not be added/updated", setName)
		assertDiff(t, dc, setName, diff)
	}
}

func assertDiff(t *testing.T, dc dirtyCacheInterface, setName string, diff testDiff) {
	if !util.IsWindowsDP() {
		if len(diff.toAdd) == 0 {
			require.Equal(t, 0, len(dc.getMembersToAdd(setName)), "expected 0 members to add")
		} else {
			require.Equal(t, stringSliceToSet(diff.toAdd), dc.getMembersToAdd(setName), "unexpected members to add for set %s", setName)
		}
		if len(diff.toDelete) == 0 {
			require.Equal(t, 0, len(dc.getMembersToDelete(setName)), "expected 0 members to delete")
		} else {
			require.Equal(t, stringSliceToSet(diff.toDelete), dc.getMembersToDelete(setName), "unexpected members to delete for set %s", setName)
		}
	}
}

func stringSliceToSet(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
