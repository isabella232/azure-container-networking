package ipsets

/*
	dirtyCacheInterface will maintain the dirty cache.
	It may maintain membersToAdd and membersToDelete.
	Members are either IPs, CIDRs, IP-Port pairs, or prefixed set names if the parent is a list.

	Assumptions:
	FIXME: TODO
*/
type dirtyCacheInterface interface {
	// reset empties dirty cache
	reset()
	// create will mark the new set to be created.
	create(set *IPSet)
	// addMember will mark the set to be updated and track the member to be added (if implemented).
	addMember(set *IPSet, member string)
	// deleteMember will mark the set to be updated and track the member to be deleted (if implemented).
	deleteMember(set *IPSet, member string)
	// delete will mark the set to be deleted in the cache
	destroy(set *IPSet)
	// getSetsToAddOrUpdate returns the set names to be added or updated
	getSetsToAddOrUpdate() map[string]struct{}
	// getSetsToDelete returns the set names to be deleted
	getSetsToDelete() map[string]struct{}
	// numSetsToAddOrUpdate returns the number of sets to be added or updated
	numSetsToAddOrUpdate() int
	// numSetsToDelete returns the number of sets to be deleted
	numSetsToDelete() int
	// isSetToAddOrUpdate returns true if the set is dirty and should be added or updated
	isSetToAddOrUpdate(setName string) bool
	// isSetToDelete returns true if the set is dirty and should be deleted
	isSetToDelete(setName string) bool
	// printAddOrUpdateCache returns a string representation of the add/update cache
	printAddOrUpdateCache() string
	// printDeleteCache returns a string representation of the delete cache
	printDeleteCache() string
	// getOriginalMembers returns the members which should be added for the set.
	getMembersToAdd(setName string) map[string]struct{}
	// getOriginalMembers returns the members which should be deleted for the set.
	getMembersToDelete(setName string) map[string]struct{}
}
