package ipsets

import "fmt"

type *memberDiff struct{}

func newMemberDiff() *memberDiff {
	return &memberDiff{}
}

func diffOnCreate(set *IPSet) *memberDiff {
	return newMemberDiff()
}

func (diff *memberDiff) addMember(member string) {
	// no-op
}

func (diff *memberDiff) deleteMember(member string) {
	// no-op
}

func (diff *memberDiff) deleteMemberIfExists(member string) {
	// no-op
}

func (diff *memberDiff) resetMembersToAdd() {
	// no-op
}
