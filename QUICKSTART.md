# 🚀 Quick Start - AWS Cost Optimization Tools

## Step 1: Install Go

```bash
sudo snap install go --classic
```

Verify installation:

```bash
go version
```

## Step 2: Build the project

```bash
cd /home/jhenriquez/repos/aws-cost-optimization-tools
./setup.sh
```

Or manually:

```bash
go mod tidy
go build -o bin/cost-optimization ./cmd/cost-optimization
```

## Step 3: Test the binary

```bash
# View help
./bin/cost-optimization --help

# View version
./bin/cost-optimization --version
```

## Step 4: Interactive Shell Mode

```bash
./bin/cost-optimization start
```

Inside the shell try:

```
aws-cost-optimization> help
aws-cost-optimization> tagging show us-east-1
aws-cost-optimization> tagging all
aws-cost-optimization> exit
```

## Step 5: CLI Mode (Examples)

### Dry-run (without applying changes)

```bash
# View resources in all regions
./bin/cost-optimization tagging show

# View resources in a specific region
./bin/cost-optimization tagging show us-east-1

# Simulate tagging in all regions
./bin/cost-optimization tagging all

# Simulate tagging in a region
./bin/cost-optimization tagging set us-east-1
```

### Apply real changes (with --apply)

```bash
# ⚠️ WARNING: These commands modify resources

# Apply tagging in all regions
./bin/cost-optimization tagging all --apply

# Apply tagging in a specific region
./bin/cost-optimization tagging set us-east-1 --apply

# Include EFS and FSx as well
./bin/cost-optimization tagging all --apply --tag-storage

# EC2 only (instances, volumes, snapshots)
./bin/cost-optimization tagging ec2 --apply

# Activate Cost Allocation Tags
./bin/cost-optimization tagging activate --apply
```

## Step 6: Install globally (optional)

```bash
sudo cp bin/cost-optimization /usr/local/bin/
cost-optimization start
```

Or with make:

```bash
make install
cost-optimization start
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

Make sure you have your AWS credentials configured:

```bash
aws configure
```

Or export environment variables:

```bash
export AWS_ACCESS_KEY_ID="your_access_key"
export AWS_SECRET_ACCESS_KEY="your_secret_key"
export AWS_REGION="us-east-1"
```

## Command Structure

### Shell Mode

```
aws-cost-optimization> tagging <mode> [options]
```

### CLI Mode

```
cost-optimization tagging <mode> [options]
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

### Example 1: Complete dry-run

```bash
cost-optimization tagging all
```

### Example 2: Apply in production

```bash
cost-optimization tagging all --apply --tag-storage
```

### Example 3: Single region only

```bash
cost-optimization tagging set us-east-1 --apply
```

### Example 4: Interactive shell

```bash
cost-optimization start

# Inside the shell:
aws-cost-optimization> tagging show
aws-cost-optimization> tagging all
aws-cost-optimization> tagging all --apply
aws-cost-optimization> exit
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
