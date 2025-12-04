package tagging

// Mode represents the operation mode for the tagging engine
type Mode string

const (
	ModeAll      Mode = "all"
	ModeSet      Mode = "set"
	ModeShow     Mode = "show"
	ModeActivate Mode = "activate"
	ModeEC2      Mode = "ec2"
	ModeEBS      Mode = "ebs"
	ModeVolumes  Mode = "volumes"
	ModeSnapshots Mode = "snapshots"
	ModeFSx      Mode = "fsx"
	ModeEFS      Mode = "efs"
	ModeDryRun   Mode = "dry-run"
)

// Options contains all configuration for the tagging engine
type Options struct {
	Mode        Mode
	Region      string
	Apply       bool
	TagStorage  bool
	FixOrphans  bool
	Regions     []string
	
	// Resource-specific flags
	TagInstances  bool
	TagVolumes    bool
	TagSnapshots  bool
	TagEFS        bool
	TagFSx        bool
}

// DefaultOptions returns options with safe defaults (dry-run mode)
func DefaultOptions() Options {
	return Options{
		Mode:         ModeAll,
		Apply:        false,
		TagStorage:   false,
		FixOrphans:   false,
		TagInstances: true,
		TagVolumes:   true,
		TagSnapshots: true,
		TagEFS:       false,
		TagFSx:       false,
	}
}
