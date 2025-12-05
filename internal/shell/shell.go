package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Th3Mayar/aws-cost-optimization-tools/internal/auth"
	"github.com/Th3Mayar/aws-cost-optimization-tools/internal/tagging"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

// Run starts the interactive shell REPL with a stylized banner and prompt.
func Run() error {
	// Print animated banner
	printBanner()

	// Build prompt with ANSI colors. 'coaws' in blue and a cyan arrow.
	blue := color.New(color.FgBlue).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	prompt := fmt.Sprintf("%s %s ", blue("coaws"), cyan("➜"))

	// History file in user's home
	histFile := "/tmp/.coaws_history"
	if home, err := os.UserHomeDir(); err == nil {
		histFile = filepath.Join(home, ".coaws_history")
	}

	// Setup autocompletion
	completer := readline.NewPrefixCompleter(
		readline.PcItem("tagging",
			readline.PcItem("all",
				readline.PcItem("--apply"),
				readline.PcItem("--tag-storage"),
				readline.PcItem("--fix-orphans"),
			),
			readline.PcItem("set",
				readline.PcItem("us-east-1"),
				readline.PcItem("us-east-2"),
				readline.PcItem("us-west-1"),
				readline.PcItem("us-west-2"),
				readline.PcItem("eu-west-1"),
				readline.PcItem("eu-central-1"),
				readline.PcItem("ap-southeast-1"),
				readline.PcItem("ap-northeast-1"),
			),
			readline.PcItem("show"),
			readline.PcItem("activate",
				readline.PcItem("--apply"),
			),
			readline.PcItem("ec2",
				readline.PcItem("--apply"),
			),
			readline.PcItem("ebs",
				readline.PcItem("--apply"),
			),
			readline.PcItem("volumes",
				readline.PcItem("--apply"),
			),
			readline.PcItem("snapshots",
				readline.PcItem("--apply"),
			),
			readline.PcItem("fsx",
				readline.PcItem("--apply"),
			),
			readline.PcItem("efs",
				readline.PcItem("--apply"),
			),
		),
		readline.PcItem("login"),
		readline.PcItem("logout"),
		readline.PcItem("whoami"),
		readline.PcItem("use-profile"),
		readline.PcItem("profiles"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("quit"),
		readline.PcItem("!clear"),
		readline.PcItem("!ls"),
		readline.PcItem("!pwd"),
		readline.PcItem("!whoami"),
	)

	rlConfig := &readline.Config{
		Prompt:          prompt,
		HistoryFile:     histFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    completer,
	}

	rl, err := readline.NewEx(rlConfig)
	if err != nil {
		return err
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				fmt.Println()
				return nil
			}
			continue
		} else if err == io.EOF {
			fmt.Println()
			return nil
		} else if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

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

		// Execute shell commands with ! prefix
		if strings.HasPrefix(line, "!") {
			shellCmd := strings.TrimPrefix(line, "!")
			if err := executeShellCommand(shellCmd); err != nil {
				fmt.Println("Error:", err)
			}
			continue
		}

		if err := handleCommand(line); err != nil {
			fmt.Println("Error:", err)
		}
	}
}

// printBanner prints a stylized ASCII banner inspired by Amazon Q
func printBanner() {
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	
	// ASCII art logo - "coaws" text
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
		fmt.Println(cyan(line))
		time.Sleep(30 * time.Millisecond)
	}

	// Info box with borders
	boxTop := "╭" + strings.Repeat("─", 78) + "╮"
	boxBottom := "╰" + strings.Repeat("─", 78) + "╯"
	
	fmt.Println(blue(boxTop))
	fmt.Println(blue("│") + white("               💰 AWS Cost Optimization & Savings Tool 💸                  ") + blue("│"))
	fmt.Println(blue("│") + white("                                                                              ") + blue("│"))
	fmt.Println(blue("│") + "     " + cyan("coaws") + white(" helps you optimize AWS costs through intelligent tagging") + "       " + blue("│"))
	fmt.Println(blue("│") + white("     and resource management across all your AWS regions.                    ") + blue("│"))
	fmt.Println(blue("│") + white("                                                                              ") + blue("│"))
	fmt.Println(blue(boxBottom))
	fmt.Println()

	// Command hints with separator
	hints := white("/help") + " all commands  •  " + 
	         white("ctrl + c") + " exit  •  " + 
	         white("↑↓") + " history  •  " +
	         white("Tab") + " autocomplete  •  " +
	         white("!cmd") + " shell"
	separator := strings.Repeat("━", 80)
	
	fmt.Println(hints)
	fmt.Println(yellow(separator))
	fmt.Println(cyan("🚀 You are using") + " " + blue("coaws") + " " + cyan("- AWS Cost Optimization Shell"))
	fmt.Println()
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("\nAuthentication:")
	fmt.Println("  login [profile]      - Configure AWS credentials")
	fmt.Println("  logout               - Clear current session")
	fmt.Println("  whoami               - Show current AWS identity")
	fmt.Println("  use-profile <name>   - Switch to different profile")
	fmt.Println("  profiles             - List available profiles")
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

// executeShellCommand runs a shell command and displays the output
func executeShellCommand(cmdStr string) error {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return nil
	}

	// Split command into parts
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return nil
	}

	// Create command
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run command
	return cmd.Run()
}
