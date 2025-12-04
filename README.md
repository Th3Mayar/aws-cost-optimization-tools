# AWS Cost Optimization Tools

FinOps mini-shell to automate tagging and cost optimization in AWS.

## Features

- 🏷️ **Tag Propagation**: Automatically propagates tags from EC2 instances to volumes, snapshots, EFS and FSx
- 💰 **Cost Allocation Tags**: Activates tags for Cost Explorer
- 🖥️ **Interactive Mode**: Shell REPL to execute commands interactively
- 🔧 **CLI Mode**: Non-interactive commands for scripts and automation
- 🔒 **Dry-Run by default**: Safe by default, requires `--apply` for real changes

## Installation

### Prerequisites

- Go 1.21 or higher
- AWS CLI configured with valid credentials

### Install Go (Ubuntu/Debian)

```bash
sudo snap install go --classic
```

### Build the binary

```bash
cd /home/jhenriquez/repos/aws-cost-optimization-tools
go mod tidy
go build -o cost-optimization ./cmd/cost-optimization
```

### Global installation (optional)

```bash
sudo mv cost-optimization /usr/local/bin/
```

## Usage

### Interactive Mode (Shell)

```bash
cost-optimization start
```

Inside the shell:

```bash
aws-cost-optimization> help
aws-cost-optimization> tagging all --apply --tag-storage
aws-cost-optimization> tagging show
aws-cost-optimization> tagging activate --apply
aws-cost-optimization> exit
```

### CLI Mode (Non-Interactive)

#### View help

```bash
cost-optimization --help
```

#### Dry-run (default, does not apply changes)

```bash
cost-optimization tagging all
cost-optimization tagging set us-east-1
cost-optimization tagging show
```

#### Apply real changes

```bash
cost-optimization tagging all --apply
cost-optimization tagging set us-east-1 --apply
cost-optimization tagging all --apply --tag-storage
```

#### Specific modes

```bash
# EC2 only (instances + volumes + snapshots)
cost-optimization tagging ec2 --apply

# EBS only (volumes + snapshots)
cost-optimization tagging ebs --apply

# Volumes only
cost-optimization tagging volumes --apply

# Snapshots only
cost-optimization tagging snapshots --apply

# FSx only
cost-optimization tagging fsx --apply

# EFS only
cost-optimization tagging efs --apply

# Activate Cost Allocation Tags
cost-optimization tagging activate --apply

# Fix orphaned snapshots
cost-optimization tagging all --apply --fix-orphans
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
