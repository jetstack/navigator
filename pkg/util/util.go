package util

// CalculateQuorum will return a quorum of the given number. This is useful
// when calculating configuration parameters for distributed systems.
func CalculateQuorum(num int32) int32 {
	if num == 0 {
		return 0
	}
	if num == 1 {
		return 1
	}
	return (num / 2) + 1
}

func Int64Ptr(i int64) *int64 {
	return &i
}
