# AWS Cost Optimization Tools - coaws

```
    ██████╗ ██████╗  █████╗ ██╗    ██╗███████╗
   ██╔════╝██╔═══██╗██╔══██╗██║    ██║██╔════╝
   ██║     ██║   ██║███████║██║ █╗ ██║███████╗
   ██║     ██║   ██║██╔══██║██║███╗██║╚════██║
   ╚██████╗╚██████╔╝██║  ██║╚███╔███╔╝███████║
    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝ ╚══╝╚══╝ ╚══════╝
```

FinOps mini-shell to automate tagging and cost optimization in AWS.

## Features

- 🔐 **AWS Authentication**: Built-in login and multi-profile support
- 🏷️ **Tag Propagation**: Automatically propagates tags from EC2 instances to volumes, snapshots, EFS and FSx
- 💰 **Cost Allocation Tags**: Activates tags for Cost Explorer
- 🖥️ **Interactive Mode**: Beautiful shell REPL with colored output and tab autocomplete
- 🔧 **CLI Mode**: Non-interactive commands for scripts and automation
- 🔒 **Dry-Run by default**: Safe by default, requires `--apply` for real changes
- 🌍 **Multi-platform**: Available for Linux, macOS, and Windows
- ⚡ **Shell Commands**: Execute system commands with `!` prefix
- 📝 **Command History**: Navigate through command history with arrow keys

## Installation

### Quick Install (Recommended)

#### Linux / macOS
```bash
curl -sSL https://raw.githubusercontent.com/Th3Mayar/aws-cost-optimization-tools/main/install.sh | bash
```

#### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/Th3Mayar/aws-cost-optimization-tools/main/install.ps1 | iex
```

### Package Managers

#### Homebrew (macOS/Linux)
```bash
brew tap Th3Mayar/coaws
brew install coaws
```

#### APT (Debian/Ubuntu)
```bash
# Download the .deb package from releases
wget https://github.com/Th3Mayar/aws-cost-optimization-tools/releases/latest/download/coaws_*_linux_amd64.deb
sudo dpkg -i coaws_*_linux_amd64.deb
```

#### YUM/DNF (RHEL/CentOS/Fedora)
```bash
# Download the .rpm package from releases
wget https://github.com/Th3Mayar/aws-cost-optimization-tools/releases/latest/download/coaws_*_linux_amd64.rpm
sudo rpm -i coaws_*_linux_amd64.rpm
```

### Manual Installation

Download the appropriate binary for your system from the [Releases page](https://github.com/Th3Mayar/aws-cost-optimization-tools/releases).

### Build from Source

#### Prerequisites
- Go 1.21 or higher
- AWS CLI configured with valid credentials

#### Build Steps
```bash
git clone https://github.com/Th3Mayar/aws-cost-optimization-tools.git
cd aws-cost-optimization-tools
go mod tidy
go build -o coaws ./cmd/cost-optimization
```

#### Global Installation (optional)
```bash
sudo mv coaws /usr/local/bin/
```

## Usage

### AWS Authentication

First, configure your AWS credentials:

```bash
coaws start

# Inside the shell:
coaws ➜ login
AWS Access Key ID: AKIA...
AWS Secret Access Key: ****
Default region [us-east-1]: us-east-1
✓ Credentials saved successfully!

# Verify your identity
coaws ➜ whoami
Current AWS Identity:
  Account: 123456789012
  User ARN: arn:aws:iam::123456789012:user/admin
  Profile: default
```

### Interactive Mode (Shell)

```bash
coaws start
```

Inside the shell:

```bash
coaws ➜ help
coaws ➜ login                    # Configure AWS credentials
coaws ➜ whoami                   # Show current identity
coaws ➜ use-profile production   # Switch profiles
coaws ➜ tagging show             # View resources
coaws ➜ tagging all --apply      # Apply tagging
coaws ➜ !clear                   # Execute shell commands
coaws ➜ exit
```

### Shell Features

- **Tab Autocomplete**: Press `Tab` to autocomplete commands and flags
- **Command History**: Use `↑` `↓` arrows to navigate history  
- **Shell Commands**: Prefix with `!` to run system commands (e.g., `!ls`, `!pwd`, `!clear`)
- **Multi-Profile**: Switch between AWS profiles without restarting

### CLI Mode (Non-Interactive)

#### View help

```bash
coaws --help
```

#### Dry-run (default, does not apply changes)

```bash
coaws tagging all
coaws tagging set us-east-1
coaws tagging show
```

#### Apply real changes

```bash
coaws tagging all --apply
coaws tagging set us-east-1 --apply
coaws tagging all --apply --tag-storage
```

#### Specific modes

```bash
# EC2 only (instances + volumes + snapshots)
coaws tagging ec2 --apply

# EBS only (volumes + snapshots)
coaws tagging ebs --apply

# Volumes only
coaws tagging volumes --apply

# Snapshots only
coaws tagging snapshots --apply

# FSx only
coaws tagging fsx --apply

# EFS only
coaws tagging efs --apply

# Activate Cost Allocation Tags
coaws tagging activate --apply

# Fix orphaned snapshots
coaws tagging all --apply --fix-orphans
```

## Project Structure

```
aws-cost-optimization-tools/
├── cmd/
│   └── cost-optimization/
│       └── main.go              # Entry point del binario
├── internal/
│   ├── tagging/
│   │   ├── options.go          # Opciones y tipos
│   │   └── engine.go           # Motor principal de tagging
│   ├── shell/
│   │   └── shell.go            # REPL interactivo
│   └── cli/
│       └── root.go             # CLI no interactivo
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Architecture

### Layers

1. **cmd/cost-optimization**: Binary entry point
2. **internal/cli**: Argument parsing and command routing
3. **internal/shell**: Interactive REPL
4. **internal/tagging**: Tagging engine (core FinOps logic)

### Execution Flow

```
main.go → cli.Run()
              ├─→ shell.Run() (modo interactivo)
              └─→ tagging.Engine.Run() (modo CLI)
```

## Security

- **Dry-run by default**: All commands run in dry-run mode unless you use `--apply`
- **No automatic changes**: Requires explicit confirmation for changes
- **AWS Credentials**: Uses the standard AWS SDK (respects `~/.aws/credentials` and environment variables)

## Supported Regions

By default processes these regions:

- us-east-1, us-east-2, us-west-1, us-west-2
- ap-south-1, ap-northeast-1/2/3, ap-southeast-1/2
- ca-central-1
- eu-central-1, eu-west-1/2/3, eu-north-1
- sa-east-1

You can specify a region with `tagging set <region>`.

## Development

### Build

```bash
make build
```

### Clean

```bash
make clean
```

### Install locally

```bash
make install
```

## Migration from Python

This project replaces the Python script `tag_propagate.py` with a complete Go implementation that offers:

- ✅ Better performance
- ✅ Single binary without dependencies
- ✅ Interactive shell mode
- ✅ Better error handling
- ✅ More maintainable code

## License

MIT License

## Author

Th3Mayar - AWS FinOps Tools
