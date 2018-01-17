package util

// CalculateQuorum will return a quorum of the given number. This is useful
// when calculating configuration parameters for distributed systems.
func CalculateQuorum(num int64) int64 {
	if num == 0 {
		return 0
	}
	if num == 1 {
		return 1
	}
	return (num / 2) + 1
}
