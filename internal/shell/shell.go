package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Th3Mayar/aws-cost-optimization-tools/internal/tagging"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

// History manages command history
type History struct {
	commands []string
	file     string
	position int
}

func newHistory(file string) *History {
	h := &History{
		commands: make([]string, 0),
		file:     file,
		position: 0,
	}
	h.load()
	return h
}

func (h *History) load() {
	data, err := os.ReadFile(h.file)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			h.commands = append(h.commands, line)
		}
	}
	h.position = len(h.commands)
}

func (h *History) save() {
	f, err := os.OpenFile(h.file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer f.Close()

	for _, cmd := range h.commands {
		fmt.Fprintln(f, cmd)
	}
}

func (h *History) add(cmd string) {
	if cmd == "" {
		return
	}
	// Avoid duplicates
	if len(h.commands) > 0 && h.commands[len(h.commands)-1] == cmd {
		return
	}
	h.commands = append(h.commands, cmd)
	h.position = len(h.commands)
	h.save()
}

// Run starts the interactive shell REPL
func Run() error {
	printBanner()

	// History file
	histFile := "/tmp/.coaws_history"
	if home, err := os.UserHomeDir(); err == nil {
		histFile = filepath.Join(home, ".coaws_history")
	}

	history := newHistory(histFile)
	reader := bufio.NewReader(os.Stdin)

	for {
		// Print prompt
		fmt.Printf("%s%scoaws%s %s➜ %s ", colorBold, colorBlue, colorReset, colorCyan, colorReset)

		// Read line
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println()
			return nil
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Add to history
		history.add(line)

		// Execute shell commands with !
		if strings.HasPrefix(line, "!") {
			shellCmd := strings.TrimPrefix(line, "!")
			if shellCmd == "" {
				continue
			}
			executeShellCommand(shellCmd)
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
			fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		}
	}
}

func printBanner() {
	// --- BIG COAWS BRAND BANNER (COMMENTED ON PURPOSE) ---
	/*
		logo := []string{
			"",
			"    ██████╗ ██████╗  █████╗ ██╗    ██╗███████╗",
			"   ██╔════╝██╔═══██╗██╔══██╗██║    ██║██╔════╝",
			"   ██║     ██║   ██║███████║██║ █╗ ██║███████╗",
			"   ██║     ██║   ██║██╔══██║██║███╗██║╚════██║",
			"   ╚██████╗╚██████╔╝██║  ██║╚███╔███╔╝███████║",
			"    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝ ╚══╝╚══╝ ╚══════╝",
			"",
		}

		// Print logo with cyan color
		for _, line := range logo {
			fmt.Printf("%s%s%s%s\n", colorBold, colorCyan, line, colorReset)
			time.Sleep(30 * time.Millisecond)
		}

		// Info box with borders
		boxTop := "╭" + strings.Repeat("─", 78) + "╮"
		boxBottom := "╰" + strings.Repeat("─", 78) + "╯"

		fmt.Printf("%s%s%s%s\n", colorBold, colorBlue, boxTop, colorReset)
		fmt.Printf("%s%s│%s%s               💰 AWS Cost Optimization & Savings Tool 💸                  %s%s│%s\n",
			colorBold, colorBlue, colorReset, colorWhite, colorBold, colorBlue, colorReset)
		fmt.Printf("%s%s│%s%s                                                                              %s%s│%s\n",
			colorBold, colorBlue, colorReset, colorWhite, colorBold, colorBlue, colorReset)
		fmt.Printf("%s%s│%s     %s%scoaws%s%s helps you optimize AWS costs through intelligent tagging%s       %s%s│%s\n",
			colorBold, colorBlue, colorReset, colorBold, colorCyan, colorReset, colorWhite, colorWhite, colorBold, colorBlue, colorReset)
		fmt.Printf("%s%s│%s%s     and resource management across all your AWS regions.                    %s%s│%s\n",
			colorBold, colorBlue, colorReset, colorWhite, colorBold, colorBlue, colorReset)
		fmt.Printf("%s%s│%s%s                                                                              %s%s│%s\n",
			colorBold, colorBlue, colorReset, colorWhite, colorBold, colorBlue, colorReset)
		fmt.Printf("%s%s%s%s\n", colorBold, colorBlue, boxBottom, colorReset)
		fmt.Println()
	*/

	// --- Minimal banner (kept) ---
	// separator := strings.Repeat("━", 80)
	// fmt.Printf("%s%s%s\n", colorYellow, separator, colorReset)
	fmt.Printf("%s🚀 AWS Cost Optimization Tool%s\n", colorCyan, colorReset)
	fmt.Println()

	// Command hints
	fmt.Printf("%shelp%s all commands  •  %sctrl + c%s exit\n",
		colorWhite, colorReset, colorWhite, colorReset)
}

func printHelp() {
	fmt.Println("Available commands:")

	// --- AUTH SECTION REMOVED ---
	/*
		fmt.Println("\nAuthentication:")
		fmt.Println("  login [profile]      - Configure AWS credentials")
		fmt.Println("  logout               - Clear current session")
		fmt.Println("  whoami               - Show current AWS identity")
		fmt.Println("  use-profile <name>   - Switch to different profile")
		fmt.Println("  profiles             - List available profiles")
	*/

	fmt.Println("\nTagging Operations:")
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

	fmt.Println("\nOther:")
	fmt.Println("  !<command>       - Execute shell command (e.g., !clear, !ls)")
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

	// --- AUTH COMMANDS REMOVED ---
	/*
		case "login":
			profile := "default"
			if len(args) > 0 {
				profile = args[0]
			}
			return auth.Login(profile)
		case "logout":
			auth.Logout()
			return nil
		case "whoami":
			return auth.Whoami()
		case "use-profile":
			if len(args) == 0 {
				fmt.Println("Usage: use-profile <profile-name>")
				return nil
			}
			return auth.UseProfile(args[0])
		case "profiles":
			profiles, err := auth.ListProfiles()
			if err != nil {
				return err
			}
			if len(profiles) == 0 {
				fmt.Println("No profiles configured. Run 'login' to create one.")
			} else {
				fmt.Println("Available profiles:")
				current := auth.GetCurrentProfile()
				for _, p := range profiles {
					if p == current {
						fmt.Printf("  * %s (active)\n", p)
					} else {
						fmt.Printf("    %s\n", p)
					}
				}
			}
			return nil
	*/

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

func executeShellCommand(cmdStr string) error {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return nil
	}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}