package shell

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Th3Mayar/aws-cost-optimization-tools/internal/tagging"
)

// Run starts the interactive shell REPL
func Run() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("AWS Cost Optimization Shell")
	fmt.Println("Type 'help' to see available commands.")
	fmt.Println("Type 'exit' or 'quit' to leave.")
	fmt.Println()

	for {
		fmt.Print("aws-cost-optimization> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "exit" || line == "quit" {
			fmt.Println("Bye.")
			return nil
		}

		if line == "help" {
			printHelp()
			continue
		}

		if err := handleCommand(line); err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  tagging all [--apply] [--tag-storage] [--fix-orphans]")
	fmt.Println("  tagging set <region> [--apply] [--tag-storage]")
	fmt.Println("  tagging show [<region>]")
	fmt.Println("  tagging activate [--apply]")
	fmt.Println("  tagging ec2 [--apply]")
	fmt.Println("  tagging ebs [--apply]")
	fmt.Println("  tagging volumes [--apply]")
	fmt.Println("  tagging snapshots [--apply]")
	fmt.Println("  tagging fsx [--apply]")
	fmt.Println("  tagging efs [--apply]")
	fmt.Println("  help")
	fmt.Println("  exit | quit")
}

func handleCommand(line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "tagging":
		return handleTagging(args)
	default:
		fmt.Println("Unknown command:", cmd)
		fmt.Println("Type 'help' for available commands.")
		return nil
	}
}

func handleTagging(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: tagging <all|set|show|activate|ec2|ebs|volumes|snapshots|fsx|efs> [options]")
		return nil
	}

	sub := args[0]
	flags := args[1:]

	opts := tagging.DefaultOptions()

	// Parse subcommand
	switch sub {
	case "all":
		opts.Mode = tagging.ModeAll
	case "set":
		opts.Mode = tagging.ModeSet
		if len(flags) > 0 && !strings.HasPrefix(flags[0], "--") {
			opts.Region = flags[0]
			flags = flags[1:]
		} else {
			fmt.Println("Error: 'set' requires a region. Usage: tagging set <region> [--apply]")
			return nil
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
	default:
		fmt.Println("Unknown tagging mode:", sub)
		fmt.Println("Available modes: all, set, show, activate, ec2, ebs, volumes, snapshots, fsx, efs")
		return nil
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
		}
	}

	// Execute
	eng := tagging.NewEngine(opts)
	return eng.Run(context.Background())
}
