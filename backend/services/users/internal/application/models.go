package application

type userInRelationToGroup struct {
	isOwner        bool
	isMember       bool
	pendingRequest bool
	pendingInvite  bool
}

type isMembershipPending struct {
	pendingRequest bool
	pendingInvite  bool
}
