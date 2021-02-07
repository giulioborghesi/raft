package domain

// roleName returns a human-readable string describing the server role
func roleName(role int) string {
	switch role {
	case candidate:
		return "candidate"
	case follower:
		return "follower"
	case leader:
		return "leader"
	default:
		panic("unknown role")
	}
}
