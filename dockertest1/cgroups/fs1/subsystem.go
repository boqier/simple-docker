package fs1

import (
	"dockertest1/cgroups/subsystem"
)

var Subsystems = []subsystem.Subsystem{
	&MemorySubsystem{},
	// &CpuSubsystem{}, // 其他子系统...
}
