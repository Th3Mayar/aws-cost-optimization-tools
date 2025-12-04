package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Th3Mayar/aws-cost-optimization-tools/internal/shell"
	"github.com/Th3Mayar/aws-cost-optimization-tools/internal/tagging"
)

// Run is the main CLI entry point
func Run(args []string) int {
	if len(args) < 2 {
		printUsage()
		return 1
	}

	cmd := args[1]

	// Handle special commands
	switch cmd {
	case "-h", "--help", "help":
		printUsage()
		return 0
	case "-v", "--version", "version":
		fmt.Printf("cost-optimization version %s\n", tagging.Version)
		return 0
	case "start":
		return runShell()
	case "tagging":
		return runTagging(args[2:])
	default:
		fmt.Println("Unknown command:", cmd)
		printUsage()
		return 1
	}
}

func printUsage() {
	fmt.Println("AWS Cost Optimization Tools")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cost-optimization start")
	fmt.Println("      Start interactive shell mode")
	fmt.Println()
	fmt.Println("  cost-optimization tagging <mode> [options]")
	fmt.Println("      Run tagging operations")
	fmt.Println()
	fmt.Println("Tagging Modes:")
	fmt.Println("  all                  Process all regions (default: dry-run)")
	fmt.Println("  set <region>         Process specific region")
	fmt.Println("  show [<region>]      Show resources only (no tagging)")
	fmt.Println("  activate             Activate Cost Allocation Tags")
	fmt.Println("  ec2                  Process only EC2 instances + volumes + snapshots")
	fmt.Println("  ebs                  Process only EBS volumes + snapshots")
	fmt.Println("  volumes              Process only EBS volumes")
	fmt.Println("  snapshots            Process only EBS snapshots")
	fmt.Println("  fsx                  Process only FSx resources")
	fmt.Println("  efs                  Process only EFS resources")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --apply              Apply changes (default: dry-run)")
	fmt.Println("  --tag-storage        Also tag EFS + FSx resources")
	fmt.Println("  --fix-orphans        Only fix orphaned AMI snapshots")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cost-optimization start")
	fmt.Println("  cost-optimization tagging all")
	fmt.Println("  cost-optimization tagging all --apply")
	fmt.Println("  cost-optimization tagging set us-east-1 --apply")
	fmt.Println("  cost-optimization tagging show")
	fmt.Println("  cost-optimization tagging activate --apply")
	fmt.Println("  cost-optimization tagging ec2 --apply")
	fmt.Println("  cost-optimization tagging all --apply --tag-storage")
}

func runShell() int {
	if err := shell.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Shell error:", err)
		return 1
	}
	return 0
}

func runTagging(args []string) int {
	if len(args) == 0 {
		fmt.Println("Error: tagging requires a mode")
		fmt.Println("Usage: cost-optimization tagging <mode> [options]")
		fmt.Println("Run 'cost-optimization --help' for more information")
		return 1
	}

	opts := tagging.DefaultOptions()
	mode := args[0]
	flags := args[1:]

	// Parse mode
	switch mode {
	case "all":
		opts.Mode = tagging.ModeAll
	case "set":
		opts.Mode = tagging.ModeSet
		if len(flags) > 0 && !strings.HasPrefix(flags[0], "--") {
			opts.Region = flags[0]
			flags = flags[1:]
		} else {
			fmt.Println("Error: 'set' requires a region")
			fmt.Println("Usage: cost-optimization tagging set <region> [--apply]")
			return 1
		}
	case "show":
		opts.Mode = tagging.ModeShow
		if len(flags) > 0 && !strings.HasPrefix(flags[0], "--") {
			opts.Region = flags[0]
			flags = flags[1:]
		}
	case "activate":
		opts.Mode = tagging.ModeActivate
	case "ec2":
		opts.Mode = tagging.ModeEC2
	case "ebs":
		opts.Mode = tagging.ModeEBS
	case "volumes":
		opts.Mode = tagging.ModeVolumes
	case "snapshots":
		opts.Mode = tagging.ModeSnapshots
	case "fsx":
		opts.Mode = tagging.ModeFSx
		opts.TagStorage = true
		opts.TagFSx = true
	case "efs":
		opts.Mode = tagging.ModeEFS
		opts.TagStorage = true
		opts.TagEFS = true
	case "dry-run":
		opts.Mode = tagging.ModeDryRun
		if len(flags) > 0 && !strings.HasPrefix(flags[0], "--") {
			opts.Region = flags[0]
			flags = flags[1:]
		}
	default:
		fmt.Println("Unknown tagging mode:", mode)
		fmt.Println("Available modes: all, set, show, activate, ec2, ebs, volumes, snapshots, fsx, efs")
		return 1
	}

	// Parse flags
	for _, f := range flags {
		switch f {
		case "--apply":
			opts.Apply = true
		case "--tag-storage":
			opts.TagStorage = true
		case "--fix-orphans":
			opts.FixOrphans = true
		default:
			fmt.Println("Unknown flag:", f)
			fmt.Println("Available flags: --apply, --tag-storage, --fix-orphans")
			return 1
		}
	}

	// Execute
	eng := tagging.NewEngine(opts)
	if err := eng.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
