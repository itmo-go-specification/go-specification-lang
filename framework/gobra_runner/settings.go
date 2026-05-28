package gobra_runner

import "time"

type Settings struct {
	JavaPath string
	GobraJAR string
	GobraBin string

	Timeout time.Duration

	Backend string

	Z3Exe string

	LogLevel string

	NoInformation             bool
	ParallelizeBranches       bool
	AssumeInjectivityOnInhale bool
	CheckOverflow             bool

	ExtraFlags []string
}

func DefaultSettings() Settings {
	return Settings{
		JavaPath:            "java",
		GobraJAR:            "gobra.jar",
		Timeout:             120 * time.Second,
		Backend:             "silicon",
		LogLevel:            "ERROR",
		NoInformation:       true,
		ParallelizeBranches: false,
	}
}
