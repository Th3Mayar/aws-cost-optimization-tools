# 🚀 Quick Start - coaws (AWS Cost Optimization Tools)

## Step 1: Installation

### Option A: Quick Install (Recommended)

#### Linux/macOS
```bash
curl -sSL https://raw.githubusercontent.com/Th3Mayar/aws-cost-optimization-tools/main/install.sh | bash
```

#### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/Th3Mayar/aws-cost-optimization-tools/main/install.ps1 | iex
```

### Option B: Build from Source

#### Install Go

```bash
sudo snap install go --classic
```

Verify installation:

```bash
go version
```

#### Build the project

```bash
cd /home/jhenriquez/repos/aws-cost-optimization-tools
./setup.sh
```

Or manually:

```bash
go mod tidy
go build -o coaws ./cmd/cost-optimization
```

#### Test the binary

```bash
# View help
./coaws --help

# View version
./coaws --version
```

## Step 2: AWS Authentication

### Configure AWS Credentials

```bash
coaws start
```

Inside the shell:

```bash
coaws ➜ login
AWS Access Key ID: AKIA...
AWS Secret Access Key: ****
Default region [us-east-1]: us-east-1
✓ Credentials saved successfully!
```

Or use existing AWS CLI profiles:

```bash
coaws ➜ use-profile production
✓ Successfully authenticated!
Account: 123456789012
User: arn:aws:iam::123456789012:user/admin
```

### Authentication Commands

```bash
coaws ➜ login [profile]      # Configure AWS credentials
coaws ➜ whoami               # Show current AWS identity
coaws ➜ use-profile <name>   # Switch to different profile
coaws ➜ profiles             # List available profiles
coaws ➜ logout               # Clear current session
```

## Step 3: Interactive Shell Mode

```bash
coaws start
```

Inside the shell try:

```bash
coaws ➜ help
coaws ➜ whoami
coaws ➜ tagging show us-east-1
coaws ➜ tagging all
coaws ➜ !ls -la              # Execute shell commands with !
coaws ➜ exit
```

### Shell Features

- **Tab Autocomplete**: Press Tab to autocomplete commands and flags
- **Command History**: Use ↑↓ arrows to navigate history
- **Shell Commands**: Prefix with `!` to execute system commands (e.g., `!clear`, `!pwd`)
- **Multi-Profile**: Switch between AWS profiles without restarting

## Step 4: CLI Mode (Examples)

### Dry-run (without applying changes)

```bash
# View resources in all regions
coaws tagging show

# View resources in a specific region
coaws tagging show us-east-1

# Simulate tagging in all regions
coaws tagging all

# Simulate tagging in a region
coaws tagging set us-east-1
```

### Apply real changes (with --apply)

```bash
# ⚠️ WARNING: These commands modify resources

# Apply tagging in all regions
coaws tagging all --apply

# Apply tagging in a specific region
coaws tagging set us-east-1 --apply

# Include EFS and FSx as well
coaws tagging all --apply --tag-storage

# EC2 only (instances, volumes, snapshots)
coaws tagging ec2 --apply

# Activate Cost Allocation Tags
coaws tagging activate --apply
```

## Step 5: Install globally (optional)

If you built from source:

```bash
sudo cp coaws /usr/local/bin/
coaws start
```

Or with make:

```bash
make install
coaws start
```

## Useful Makefile Commands

```bash
make help          # View all available commands
make build         # Build the binary
make install       # Install globally
make clean         # Clean compiled files
make run           # Run in interactive mode
make test-dry-run  # Test in dry-run mode
make test-show     # Show resources
```

## AWS Configuration

### Option 1: Using coaws (Recommended)

```bash
coaws start
coaws ➜ login
```

### Option 2: AWS CLI profiles

```bash
aws configure --profile myprofile
```

Then use it in coaws:

```bash
coaws start
coaws ➜ use-profile myprofile
```

### Option 3: Environment variables

```bash
export AWS_ACCESS_KEY_ID="your_access_key"
export AWS_SECRET_ACCESS_KEY="your_secret_key"
export AWS_REGION="us-east-1"
coaws start
```

## Command Structure

### Shell Mode

```bash
coaws ➜ tagging <mode> [options]
coaws ➜ login [profile]
coaws ➜ use-profile <name>
coaws ➜ whoami
coaws ➜ !<shell-command>
```

### CLI Mode

```bash
coaws tagging <mode> [options]
coaws start  # Interactive shell
coaws --help
coaws --version
```

### Available modes

- `all` - All regions
- `set <region>` - Specific region
- `show [region]` - View resources (without modifying)
- `activate` - Activate Cost Allocation Tags
- `ec2` - EC2 + EBS only
- `ebs` - EBS only (volumes + snapshots)
- `volumes` - EBS volumes only
- `snapshots` - EBS snapshots only
- `fsx` - FSx only
- `efs` - EFS only

### Options (flags)

- `--apply` - Apply changes (default: dry-run)
- `--tag-storage` - Include EFS and FSx
- `--fix-orphans` - Fix orphaned snapshots

## Complete Examples

### Example 1: First time setup and dry-run

```bash
coaws start

# Configure credentials
coaws ➜ login
AWS Access Key ID: AKIA...
AWS Secret Access Key: ****
Default region: us-east-1

# Verify identity
coaws ➜ whoami

# Test in dry-run
coaws ➜ tagging all
coaws ➜ exit
```

### Example 2: Apply in production

```bash
coaws tagging all --apply --tag-storage
```

### Example 3: Multi-profile workflow

```bash
coaws start

# Switch to production profile
coaws ➜ use-profile production
✓ Successfully authenticated!

# Apply changes
coaws ➜ tagging all --apply

# Switch to staging
coaws ➜ use-profile staging
coaws ➜ tagging all --apply

coaws ➜ exit
```

### Example 4: Single region only

```bash
coaws tagging set us-east-1 --apply
```

### Example 5: Interactive shell with utilities

```bash
coaws start

# Inside the shell:
coaws ➜ whoami
coaws ➜ !pwd
coaws ➜ tagging show
coaws ➜ tagging all
coaws ➜ !clear
coaws ➜ tagging all --apply
coaws ➜ exit
```

## ⚠️ Warnings

- **Always test first in dry-run** (without `--apply`)
- Changes with `--apply` are **PERMANENT**
- Review required IAM permissions
- Save execution logs for auditing

## Required IAM Permissions

Your AWS user/role needs these permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeVolumes",
        "ec2:DescribeSnapshots",
        "ec2:DescribeTags",
        "ec2:CreateTags",
        "efs:DescribeFileSystems",
        "efs:DescribeAccessPoints",
        "efs:ListTagsForResource",
        "efs:TagResource",
        "fsx:DescribeFileSystems",
        "fsx:DescribeBackups",
        "fsx:DescribeVolumes",
        "fsx:ListTagsForResource",
        "fsx:TagResource",
        "ce:ListCostAllocationTags",
        "ce:UpdateCostAllocationTagsStatus"
      ],
      "Resource": "*"
    }
  ]
}
```

## Troubleshooting

### Error: "Go is not installed"

```bash
sudo snap install go --classic
```

### Error: "cannot find module"

```bash
go mod tidy
```

### Error: "access denied"

Verify your AWS credentials:

```bash
aws sts get-caller-identity
```

### Binary does not execute

```bash
chmod +x bin/cost-optimization
```

## Next Steps

1. ✅ Install Go
2. ✅ Build the project
3. ✅ Test in dry-run
4. ✅ Review output
5. ✅ Apply changes with `--apply`
6. 📊 Monitor in AWS Console
7. 💰 Review Cost Explorer (24-48h later)

## Support

- Read the complete README.md
- Review planning.md to understand the architecture
- Report issues in the repository

---

**Ready to optimize costs! 🚀💰**
