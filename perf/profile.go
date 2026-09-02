package perf

import (
	"fmt"
	"os"
	"runtime/pprof"
)

func StartProfileCPU() func() {
	cpuFile, err := os.Create("cpu.pprof")
	if err != nil {
		panic(fmt.Errorf("create CPU profile: %w", err))
	}

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		panic(fmt.Errorf("start CPU profile: %w", err))
	}

	return func() {
		pprof.StopCPUProfile()
		cpuFile.Close()
	}
}
