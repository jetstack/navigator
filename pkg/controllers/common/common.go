package common

const (
	JVMCgroupOpts =
	// allow using cgroup flags
	"-XX:+UnlockExperimentalVMOptions " +

		// use cgroup limit instead of host
		"-XX:+UseCGroupMemoryLimitForHeap " +

		// use 100% of the available memory
		"-XX:MaxRAMFraction=1 "
)
