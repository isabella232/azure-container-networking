package ipsets

import (
	"testing"

	"github.com/Azure/azure-container-networking/npm/util"
	"github.com/stretchr/testify/require"
)

// members are only important for Linux
type dirtyCacheResults struct {
	// map of prefixed name to members
	toAddOrUpdate map[string][]string
	// map of prefixed name to members
	toDelete map[string][]string
}

const (
	ip     = "1.2.3.4"
	podKey = "pod1"
)

func TestDirtyCacheReset(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	set2 := NewIPSet(NewIPSetMetadata("set2", Namespace))
	dc := newDirtyCache()
	set1.IPPodKey[ip] = podKey
	dc.create(set1)
	dc.delete(set2)
	dc.reset()
	assertDirtyCache(t, dc, &dirtyCacheResults{})
}

func TestDirtyCacheCreate(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	set2 := NewIPSet(NewIPSetMetadata("set2", Namespace))
	dc := newDirtyCache()
	dc.create(set1)
	// members are ignored on create
	set2.IPPodKey[ip] = podKey
	dc.create(set2)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: map[string][]string{
			set1.Name: {},
			set2.Name: {},
		},
		toDelete: nil,
	})
}

func TestDirtyCacheCreateAfterDelete(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	dc := newDirtyCache()
	set1.IPPodKey[ip] = podKey
	dc.delete(set1)
	// original members shouldn't get updated
	dc.create(set1)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: map[string][]string{
			set1.Name: {ip},
		},
		toDelete: nil,
	})
}

func TestDirtyCacheUpdateNew(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	dc := newDirtyCache()
	set1.IPPodKey[ip] = podKey
	dc.update(set1)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: map[string][]string{
			set1.Name: {ip},
		},
		toDelete: nil,
	})
}

func TestDirtyCacheUpdateOld(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	dc := newDirtyCache()
	dc.create(set1)
	// original members shouldn't get updated
	set1.IPPodKey[ip] = podKey
	dc.update(set1)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: map[string][]string{
			set1.Name: {},
		},
		toDelete: nil,
	})
}

func TestDirtyCacheUpdateTwice(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	dc := newDirtyCache()
	dc.update(set1)
	// original members shouldn't get updated
	set1.IPPodKey[ip] = podKey
	dc.update(set1)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: map[string][]string{
			set1.Name: {},
		},
		toDelete: nil,
	})
}

func TestDirtyCacheUpdateAfterDelete(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	dc := newDirtyCache()
	dc.delete(set1)
	// original members shouldn't get updated
	set1.IPPodKey[ip] = podKey
	dc.update(set1)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: map[string][]string{
			set1.Name: {},
		},
		toDelete: nil,
	})
}

func TestDirtyCacheDelete(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	dc := newDirtyCache()
	set1.IPPodKey[ip] = podKey
	dc.delete(set1)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: nil,
		toDelete: map[string][]string{
			set1.Name: {ip},
		},
	})
}

func TestDirtyCacheDeleteOld(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	dc := newDirtyCache()
	set1.IPPodKey[ip] = podKey
	dc.update(set1)
	// original members shouldn't get updated
	delete(set1.IPPodKey, ip)
	dc.delete(set1)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: nil,
		toDelete: map[string][]string{
			set1.Name: {ip},
		},
	})
}

func TestDirtyCacheDeleteTwice(t *testing.T) {
	set1 := NewIPSet(NewIPSetMetadata("set1", Namespace))
	dc := newDirtyCache()
	set1.IPPodKey[ip] = podKey
	dc.update(set1)
	// original members shouldn't get updated
	delete(set1.IPPodKey, ip)
	dc.delete(set1)
	assertDirtyCache(t, dc, &dirtyCacheResults{
		toAddOrUpdate: nil,
		toDelete: map[string][]string{
			set1.Name: {ip},
		},
	})
}

func assertDirtyCache(t *testing.T, dc dirtyCacheMaintainer, expected *dirtyCacheResults) {
	require.Equal(t, len(expected.toAddOrUpdate), dc.numSetsToAddOrUpdate(), "unexpected number of sets to add or update")
	require.Equal(t, len(expected.toDelete), dc.numSetsToDelete(), "unexpected number of sets to delete")
	for setName, members := range expected.toAddOrUpdate {
		require.True(t, dc.isSetToAddOrUpdate(setName), "set %s should be added/updated", setName)
		require.False(t, dc.isSetToDelete(setName), "set %s should not be deleted", setName)
		if !util.IsWindowsDP() {
			require.Equal(t, stringSliceToSet(members), dc.getOriginalMembers(setName), "unexpected original members for set %s", setName)
		}
	}
	for setName, members := range expected.toDelete {
		require.True(t, dc.isSetToDelete(setName), "set %s should be deleted", setName)
		require.False(t, dc.isSetToAddOrUpdate(setName), "set %s should not be added/updated", setName)
		if !util.IsWindowsDP() {
			require.Equal(t, stringSliceToSet(members), dc.getOriginalMembers(setName), "unexpected original members for set %s", setName)
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
