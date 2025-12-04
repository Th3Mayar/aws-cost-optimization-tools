package tagging

import "testing"

func TestDefaultOptions_SafeDefaults(t *testing.T) {
	opts := DefaultOptions()

	if opts.Mode != ModeAll {
		t.Errorf("expected default Mode to be %q, got %q", ModeAll, opts.Mode)
	}

	if opts.Apply {
		t.Errorf("expected Apply to be false by default")
	}

	if opts.TagStorage {
		t.Errorf("expected TagStorage to be false by default")
	}

	if opts.FixOrphans {
		t.Errorf("expected FixOrphans to be false by default")
	}

	// Resource flags
	if !opts.TagInstances {
		t.Errorf("expected TagInstances to be true by default")
	}
	if !opts.TagVolumes {
		t.Errorf("expected TagVolumes to be true by default")
	}
	if !opts.TagSnapshots {
		t.Errorf("expected TagSnapshots to be true by default")
	}
	if opts.TagEFS {
		t.Errorf("expected TagEFS to be false by default")
	}
	if opts.TagFSx {
		t.Errorf("expected TagFSx to be false by default")
	}
}

func TestMode_StringValues(t *testing.T) {
	cases := []struct {
		mode Mode
		want string
	}{
		{ModeAll, "all"},
		{ModeSet, "set"},
		{ModeShow, "show"},
		{ModeActivate, "activate"},
		{ModeEC2, "ec2"},
		{ModeEBS, "ebs"},
		{ModeVolumes, "volumes"},
		{ModeSnapshots, "snapshots"},
		{ModeFSx, "fsx"},
		{ModeEFS, "efs"},
		{ModeDryRun, "dry-run"},
	}

	for _, tc := range cases {
		if string(tc.mode) != tc.want {
			t.Errorf("expected mode %q to have string value %q, got %q", tc.mode, tc.want, string(tc.mode))
		}
	}
}