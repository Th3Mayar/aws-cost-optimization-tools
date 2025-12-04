package tagging

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	types2 "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	fsxtypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
)

const Version = "0.1.5"

var TargetRegions = []string{
	"us-east-1", "us-east-2", "us-west-1", "us-west-2",
	"ap-south-1", "ap-northeast-3", "ap-northeast-2",
	"ap-southeast-1", "ap-southeast-2", "ap-northeast-1",
	"ca-central-1", "eu-central-1", "eu-west-1",
	"eu-west-2", "eu-west-3", "eu-north-1", "sa-east-1",
}

// Engine is the main tagging engine
type Engine struct {
	opts Options
	cfg  aws.Config
}

// NewEngine creates a new tagging engine with the given options
func NewEngine(opts Options) *Engine {
	return &Engine{opts: opts}
}

// Run executes the tagging operation based on the configured mode
func (e *Engine) Run(ctx context.Context) error {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	e.cfg = cfg

	// Determine regions to process
	regions := e.resolveRegions(ctx)
	if len(regions) == 0 {
		return fmt.Errorf("no regions to process")
	}

	// Execute based on mode
	switch e.opts.Mode {
	case ModeShow:
		return e.runShow(ctx, regions)
	case ModeActivate:
		return e.runActivate(ctx, regions)
	case ModeEC2:
		return e.runEC2(ctx, regions)
	case ModeEBS:
		return e.runEBS(ctx, regions)
	case ModeVolumes:
		return e.runVolumes(ctx, regions)
	case ModeSnapshots:
		return e.runSnapshots(ctx, regions)
	case ModeFSx:
		return e.runFSx(ctx, regions)
	case ModeEFS:
		return e.runEFSOnly(ctx, regions)
	case ModeAll, ModeSet, ModeDryRun:
		return e.runAllResources(ctx, regions)
	default:
		return fmt.Errorf("unknown mode: %s", e.opts.Mode)
	}
}

// resolveRegions determines which regions to operate on
func (e *Engine) resolveRegions(ctx context.Context) []string {
	if len(e.opts.Regions) > 0 {
		return e.opts.Regions
	}

	if e.opts.Region != "" {
		return []string{e.opts.Region}
	}

	// Use target regions
	if len(TargetRegions) > 0 {
		return TargetRegions
	}

	// Fallback: describe all regions
	ec2Client := ec2.NewFromConfig(e.cfg)
	result, err := ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		fmt.Printf("[WARN] Failed to describe regions: %v\n", err)
		return []string{"us-east-1"} // Fallback
	}

	regions := make([]string, len(result.Regions))
	for i, r := range result.Regions {
		regions[i] = aws.ToString(r.RegionName)
	}
	return regions
}

// runShow lists resources without modifying anything
func (e *Engine) runShow(ctx context.Context, regions []string) error {
	for _, region := range regions {
		e.showRegion(ctx, region)
	}
	return nil
}

func (e *Engine) showRegion(ctx context.Context, region string) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("[SHOW] REGION: %s\n", strings.ToUpper(region))
	fmt.Printf("%s\n", strings.Repeat("=", 80))

	regionCfg := e.cfg.Copy()
	regionCfg.Region = region

	// EC2 instances
	ec2Client := ec2.NewFromConfig(regionCfg)
	instances, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "stopped"},
			},
		},
	})
	if err == nil {
		count := 0
		for _, res := range instances.Reservations {
			count += len(res.Instances)
		}
		fmt.Printf("[EC2] Instances: %d\n", count)
	}

	// EFS
	efsClient := efs.NewFromConfig(regionCfg)
	fsResult, err := efsClient.DescribeFileSystems(ctx, &efs.DescribeFileSystemsInput{})
	if err == nil {
		fmt.Printf("[EFS] FileSystems: %d\n", len(fsResult.FileSystems))
	} else {
		fmt.Println("[EFS] Not accessible or no EFS in this region")
	}

	// FSx
	fsxClient := fsx.NewFromConfig(regionCfg)
	fsxResult, err := fsxClient.DescribeFileSystems(ctx, &fsx.DescribeFileSystemsInput{})
	if err == nil {
		fmt.Printf("[FSx] FileSystems: %d\n", len(fsxResult.FileSystems))
	} else {
		fmt.Println("[FSx] Not accessible or no FSx in this region")
	}
}

// runActivate activates cost allocation tags
func (e *Engine) runActivate(ctx context.Context, regions []string) error {
	mode := "DRY-RUN"
	if e.opts.Apply {
		mode = "APPLY"
	}

	fmt.Printf("\n[COST ALLOCATION TAGS] Activating eligible tag keys\n")
	fmt.Printf("Regions: %s\n", strings.Join(regions, ", "))
	fmt.Printf("Mode: %s\n\n", mode)

	ceClient := costexplorer.NewFromConfig(e.cfg)

	// List current cost allocation tags
	current, err := ceClient.ListCostAllocationTags(ctx, &costexplorer.ListCostAllocationTagsInput{})
	if err != nil {
		return fmt.Errorf("cannot list Cost Allocation Tags: %w", err)
	}

	activeKeys := make(map[string]bool)
	for _, tag := range current.CostAllocationTags {
		activeKeys[aws.ToString(tag.TagKey)] = true
	}
	fmt.Printf("Currently active Cost Allocation Tags: %d\n", len(activeKeys))

	// Collect all tag keys from all regions
	allKeys := make(map[string]bool)
	for _, region := range regions {
		fmt.Printf("  Scanning region %s...\n", strings.ToUpper(region))
		regionCfg := e.cfg.Copy()
		regionCfg.Region = region
		ec2Client := ec2.NewFromConfig(regionCfg)

		result, err := ec2Client.DescribeTags(ctx, &ec2.DescribeTagsInput{})
		if err == nil {
			for _, tag := range result.Tags {
				key := aws.ToString(tag.Key)
				if key != "" {
					allKeys[key] = true
				}
			}
		}
	}

	// Find eligible keys
	eligible := []string{}
	for key := range allKeys {
		if !activeKeys[key] {
			eligible = append(eligible, key)
		}
	}

	fmt.Printf("\nFound %d unique tag keys → %d eligible for activation\n", len(allKeys), len(eligible))

	if len(eligible) == 0 {
		fmt.Println("No new Cost Allocation Tags to activate.")
		return nil
	}

	fmt.Println("\nTag keys to activate:")
	status := "PLAN"
	if e.opts.Apply {
		status = "APPLY"
	}
	for _, key := range eligible {
		fmt.Printf("    [%s] %s\n", status, key)
	}

	if !e.opts.Apply {
		fmt.Println("\nDRY-RUN: No changes made. Use --apply to activate.")
		return nil
	}

	// Activate tags
	_, err = ceClient.UpdateCostAllocationTagsStatus(ctx, &costexplorer.UpdateCostAllocationTagsStatusInput{
		CostAllocationTagsStatus: buildCostAllocationTagStatus(eligible),
	})
	if err != nil {
		return fmt.Errorf("failed to activate Cost Allocation Tags: %w", err)
	}

	fmt.Printf("\nSUCCESS: %d Cost Allocation Tags activated!\n", len(eligible))
	fmt.Println("Cost Explorer will reflect these tags within 24-48 hours.")
	return nil
}

func buildCostAllocationTagStatus(keys []string) []types2.CostAllocationTagStatusEntry {
	entries := make([]types2.CostAllocationTagStatusEntry, len(keys))
	for i, key := range keys {
		entries[i] = types2.CostAllocationTagStatusEntry{
			TagKey: aws.String(key),
			Status: types2.CostAllocationTagStatusActive,
		}
	}
	return entries
}

// runAllResources processes EC2 instances and optionally storage resources
func (e *Engine) runAllResources(ctx context.Context, regions []string) error {
	mode := "DRY-RUN MODE"
	if e.opts.Apply {
		mode = "APPLY MODE – REAL CHANGES!"
	}

	fmt.Printf("\n%s\n", mode)
	fmt.Printf("Action: %s\n", e.opts.Mode)
	fmt.Printf("Target regions: %s\n\n", strings.Join(regions, ", "))

	for _, region := range regions {
		if e.opts.FixOrphans {
			e.fixOrphanedSnapshots(ctx, region)
		} else {
			e.processRegion(ctx, region)
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("═", 80))
	fmt.Println("TAG PROPAGATION COMPLETED!")
	if e.opts.TagStorage {
		fmt.Println("EC2 + EFS + FSx resources were processed.")
	} else {
		fmt.Println("EC2 resources were processed. Use --tag-storage to include EFS/FSx.")
	}
	fmt.Printf("%s\n", strings.Repeat("═", 80))

	return nil
}

// runEC2 processes only EC2 instances, volumes, and snapshots
func (e *Engine) runEC2(ctx context.Context, regions []string) error {
	e.opts.TagInstances = true
	e.opts.TagVolumes = true
	e.opts.TagSnapshots = true
	e.opts.TagStorage = false
	return e.runAllResources(ctx, regions)
}

// runEBS processes only EBS volumes and snapshots
func (e *Engine) runEBS(ctx context.Context, regions []string) error {
	for _, region := range regions {
		e.processAllVolumes(ctx, region)
		e.processAllSnapshots(ctx, region)
	}
	return nil
}

// runVolumes processes only EBS volumes
func (e *Engine) runVolumes(ctx context.Context, regions []string) error {
	for _, region := range regions {
		e.processAllVolumes(ctx, region)
	}
	return nil
}

// runSnapshots processes only EBS snapshots
func (e *Engine) runSnapshots(ctx context.Context, regions []string) error {
	for _, region := range regions {
		e.processAllSnapshots(ctx, region)
	}
	return nil
}

// runFSx processes only FSx resources
func (e *Engine) runFSx(ctx context.Context, regions []string) error {
	for _, region := range regions {
		e.processFSx(ctx, region)
	}
	return nil
}

// runEFSOnly processes only EFS resources
func (e *Engine) runEFSOnly(ctx context.Context, regions []string) error {
	for _, region := range regions {
		e.processEFS(ctx, region)
	}
	return nil
}

// processRegion processes all resources in a single region
func (e *Engine) processRegion(ctx context.Context, region string) {
	mode := "DRY-RUN"
	if e.opts.Apply {
		mode = "APPLY"
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("REGION: %s | Mode: %s\n", strings.ToUpper(region), mode)
	fmt.Printf("%s\n", strings.Repeat("=", 80))

	regionCfg := e.cfg.Copy()
	regionCfg.Region = region

	ec2Client := ec2.NewFromConfig(regionCfg)

	// Process EC2 instances
	result, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "stopped"},
			},
		},
	})
	if err != nil {
		fmt.Printf("[ERROR] Failed to describe instances in %s: %v\n", region, err)
		return
	}

	count := 0
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			e.processInstance(ctx, ec2Client, instance)
			count++
		}
	}

	fmt.Printf("[SUMMARY] %s → %d instances processed\n", region, count)

	// Process storage if requested
	if e.opts.TagStorage {
		e.processEFS(ctx, region)
		e.processFSx(ctx, region)
	}
}

// processInstance processes a single EC2 instance and its volumes/snapshots
func (e *Engine) processInstance(ctx context.Context, client *ec2.Client, instance types.Instance) {
	if instance.State.Name == types.InstanceStateNameTerminated {
		return
	}

	machineKey := e.getMachineKey(instance)
	nameValue := e.getNameTag(instance)
	if nameValue == "" {
		nameValue = aws.ToString(instance.InstanceId)
	}

	display := nameValue
	if nameValue != aws.ToString(instance.InstanceId) {
		display = fmt.Sprintf("%s (%s)", nameValue, aws.ToString(instance.InstanceId))
	}

	fmt.Printf("\n[PROCESSING] %s → Using tag key: '%s'\n", display, machineKey)

	// Tag instance itself
	if e.opts.TagInstances {
		currentTags := make(map[string]string)
		for _, tag := range instance.Tags {
			currentTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
		}

		tagsToAdd := []types.Tag{}
		if _, exists := currentTags["Name"]; !exists {
			tagsToAdd = append(tagsToAdd, types.Tag{Key: aws.String("Name"), Value: aws.String(nameValue)})
		}
		if _, exists := currentTags[machineKey]; !exists {
			tagsToAdd = append(tagsToAdd, types.Tag{Key: aws.String(machineKey), Value: aws.String("")})
		}

		e.planOrApply(ctx, client, aws.ToString(instance.InstanceId), tagsToAdd, "EC2 Instance")
	}

	// Tag volumes and snapshots
	if e.opts.TagVolumes || e.opts.TagSnapshots {
		e.tagVolumesAndSnapshots(ctx, client, instance, machineKey, nameValue)
	}
}

// getMachineKey returns the appropriate tag key for the instance
func (e *Engine) getMachineKey(instance types.Instance) string {
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == "Name" {
			name := aws.ToString(tag.Value)
			normalized := normalizeKey(name)
			if normalized != "" {
				return normalized
			}
		}
	}
	return aws.ToString(instance.InstanceId)
}

// getNameTag returns the Name tag value
func (e *Engine) getNameTag(instance types.Instance) string {
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == "Name" {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

// normalizeKey normalizes a tag key
func normalizeKey(s string) string {
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(`[^a-zA-Z0-9\-\s]`)
	s = re.ReplaceAllString(s, "-")
	return strings.ReplaceAll(s, " ", "-")
}

// tagVolumesAndSnapshots tags volumes and snapshots associated with an instance
func (e *Engine) tagVolumesAndSnapshots(ctx context.Context, client *ec2.Client, instance types.Instance, machineKey, nameValue string) {
	volumeIDs := []string{}

	// Collect volume IDs
	for _, mapping := range instance.BlockDeviceMappings {
		if mapping.Ebs != nil && mapping.Ebs.VolumeId != nil {
			volID := aws.ToString(mapping.Ebs.VolumeId)
			volumeIDs = append(volumeIDs, volID)

			if e.opts.TagVolumes {
				e.processResource(ctx, client, volID, machineKey, nameValue, "Volume")
			}
		}
	}

	if !e.opts.TagSnapshots {
		return
	}

	// Tag snapshots from volumes
	if len(volumeIDs) > 0 {
		paginator := ec2.NewDescribeSnapshotsPaginator(client, &ec2.DescribeSnapshotsInput{
			OwnerIds: []string{"self"},
			Filters: []types.Filter{
				{Name: aws.String("volume-id"), Values: volumeIDs},
			},
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				fmt.Printf("    [WARN] Failed to describe snapshots: %v\n", err)
				break
			}
			for _, snapshot := range page.Snapshots {
				e.processResource(ctx, client, aws.ToString(snapshot.SnapshotId), machineKey, nameValue, "Snapshot")
			}
		}
	}

	// Tag AMI snapshots (by instance ID in description)
	instanceID := aws.ToString(instance.InstanceId)
	paginator := ec2.NewDescribeSnapshotsPaginator(client, &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			break
		}
		for _, snapshot := range page.Snapshots {
			desc := aws.ToString(snapshot.Description)
			if strings.Contains(desc, instanceID) {
				e.processResource(ctx, client, aws.ToString(snapshot.SnapshotId), machineKey, nameValue, "Snapshot")
			}
		}
	}
}

// processResource processes a single EC2 resource (volume or snapshot)
func (e *Engine) processResource(ctx context.Context, client *ec2.Client, resourceID, machineKey, nameValue, resourceType string) {
	var currentTags map[string]string

	if resourceType == "Volume" {
		result, err := client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
			VolumeIds: []string{resourceID},
		})
		if err != nil || len(result.Volumes) == 0 {
			return
		}
		currentTags = make(map[string]string)
		for _, tag := range result.Volumes[0].Tags {
			currentTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
		}
	} else {
		result, err := client.DescribeSnapshots(ctx, &ec2.DescribeSnapshotsInput{
			SnapshotIds: []string{resourceID},
		})
		if err != nil || len(result.Snapshots) == 0 {
			return
		}
		currentTags = make(map[string]string)
		for _, tag := range result.Snapshots[0].Tags {
			currentTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
		}
	}

	tagsToAdd := []types.Tag{}
	if _, exists := currentTags["Name"]; !exists {
		tagsToAdd = append(tagsToAdd, types.Tag{Key: aws.String("Name"), Value: aws.String(nameValue)})
	}
	if _, exists := currentTags[machineKey]; !exists {
		tagsToAdd = append(tagsToAdd, types.Tag{Key: aws.String(machineKey), Value: aws.String("")})
	}

	e.planOrApply(ctx, client, resourceID, tagsToAdd, resourceType)
}

// planOrApply either plans or applies tags to a resource
func (e *Engine) planOrApply(ctx context.Context, client *ec2.Client, resourceID string, tags []types.Tag, resourceType string) {
	if len(tags) == 0 {
		return
	}

	action := "PLAN"
	if e.opts.Apply {
		action = "APPLY"
	}

	for _, tag := range tags {
		value := aws.ToString(tag.Value)
		if value == "" {
			value = "(empty)"
		}
		fmt.Printf("    [%s] %s %s → %s = %s\n", action, resourceType, resourceID, aws.ToString(tag.Key), value)
	}

	if !e.opts.Apply {
		return
	}

	_, err := client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{resourceID},
		Tags:      tags,
	})
	if err != nil {
		fmt.Printf("    [ERROR] %s %s: %v\n", resourceType, resourceID, err)
	}
}

// processAllVolumes processes all EBS volumes in a region
func (e *Engine) processAllVolumes(ctx context.Context, region string) {
	mode := "DRY-RUN"
	if e.opts.Apply {
		mode = "APPLY"
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("REGION: %s | Mode: %s\n", strings.ToUpper(region), mode)
	fmt.Printf("%s\n", strings.Repeat("=", 80))

	regionCfg := e.cfg.Copy()
	regionCfg.Region = region
	client := ec2.NewFromConfig(regionCfg)

	fmt.Println("\n[VOLUMES MODE] Processing all EBS volumes...")
	paginator := ec2.NewDescribeVolumesPaginator(client, &ec2.DescribeVolumesInput{})

	count := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			fmt.Printf("[ERROR] Failed to describe volumes: %v\n", err)
			break
		}

		for _, volume := range page.Volumes {
			volumeID := aws.ToString(volume.VolumeId)
			
			currentTags := make(map[string]string)
			for _, tag := range volume.Tags {
				currentTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}

			tagsToAdd := []types.Tag{}
			nameValue := currentTags["Name"]
			if nameValue == "" {
				nameValue = volumeID
				tagsToAdd = append(tagsToAdd, types.Tag{Key: aws.String("Name"), Value: aws.String(nameValue)})
			}

			machineKey := normalizeKey(nameValue)
			if machineKey == "" {
				machineKey = volumeID
			}
			if _, exists := currentTags[machineKey]; !exists {
				tagsToAdd = append(tagsToAdd, types.Tag{Key: aws.String(machineKey), Value: aws.String("")})
			}

			e.planOrApply(ctx, client, volumeID, tagsToAdd, "Volume")
			count++
		}
	}

	fmt.Printf("\n[SUMMARY] %s → %d volumes processed\n", region, count)
}

// processAllSnapshots processes all EBS snapshots in a region
func (e *Engine) processAllSnapshots(ctx context.Context, region string) {
	mode := "DRY-RUN"
	if e.opts.Apply {
		mode = "APPLY"
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("REGION: %s | Mode: %s\n", strings.ToUpper(region), mode)
	fmt.Printf("%s\n", strings.Repeat("=", 80))

	regionCfg := e.cfg.Copy()
	regionCfg.Region = region
	client := ec2.NewFromConfig(regionCfg)

	fmt.Println("\n[SNAPSHOTS MODE] Processing all EBS snapshots...")
	paginator := ec2.NewDescribeSnapshotsPaginator(client, &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	})

	count := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			fmt.Printf("[ERROR] Failed to describe snapshots: %v\n", err)
			break
		}

		for _, snapshot := range page.Snapshots {
			snapshotID := aws.ToString(snapshot.SnapshotId)
			
			currentTags := make(map[string]string)
			for _, tag := range snapshot.Tags {
				currentTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}

			tagsToAdd := []types.Tag{}
			nameValue := currentTags["Name"]
			if nameValue == "" {
				nameValue = snapshotID
				tagsToAdd = append(tagsToAdd, types.Tag{Key: aws.String("Name"), Value: aws.String(nameValue)})
			}

			machineKey := normalizeKey(nameValue)
			if machineKey == "" {
				machineKey = snapshotID
			}
			if _, exists := currentTags[machineKey]; !exists {
				tagsToAdd = append(tagsToAdd, types.Tag{Key: aws.String(machineKey), Value: aws.String("")})
			}

			e.planOrApply(ctx, client, snapshotID, tagsToAdd, "Snapshot")
			count++
		}
	}

	fmt.Printf("\n[SUMMARY] %s → %d snapshots processed\n", region, count)
}

// fixOrphanedSnapshots fixes orphaned AMI snapshots
func (e *Engine) fixOrphanedSnapshots(ctx context.Context, region string) {
	fmt.Println("\n[ORPHAN MODE] Fixing orphaned AMI snapshots that have no Name tag...")
	
	regionCfg := e.cfg.Copy()
	regionCfg.Region = region
	client := ec2.NewFromConfig(regionCfg)

	paginator := ec2.NewDescribeSnapshotsPaginator(client, &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	})

	fixed := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			fmt.Printf("[ERROR] Failed to describe snapshots: %v\n", err)
			break
		}

		for _, snapshot := range page.Snapshots {
			hasName := false
			for _, tag := range snapshot.Tags {
				if aws.ToString(tag.Key) == "Name" {
					hasName = true
					break
				}
			}

			if !hasName {
				desc := aws.ToString(snapshot.Description)
				if strings.Contains(desc, "Created by CreateImage") {
					nameValue := fmt.Sprintf("AMI-Snapshot-%s", aws.ToString(snapshot.SnapshotId))
					tags := []types.Tag{
						{Key: aws.String("Name"), Value: aws.String(nameValue)},
					}
					e.planOrApply(ctx, client, aws.ToString(snapshot.SnapshotId), tags, "Orphaned Snapshot")
					fixed++
				}
			}
		}
	}

	fmt.Printf("\n[ORPHAN MODE] Completed → %d orphaned AMI snapshots fixed!\n", fixed)
}

// processEFS processes EFS resources in a region
func (e *Engine) processEFS(ctx context.Context, region string) {
	mode := "DRY-RUN"
	if e.opts.Apply {
		mode = "APPLY"
	}

	fmt.Printf("\n[EFS] Processing EFS resources in %s (%s)\n", strings.ToUpper(region), mode)

	regionCfg := e.cfg.Copy()
	regionCfg.Region = region
	client := efs.NewFromConfig(regionCfg)

	// File Systems
	fsResult, err := client.DescribeFileSystems(ctx, &efs.DescribeFileSystemsInput{})
	if err != nil {
		fmt.Printf("[WARN] Failed to describe EFS file systems: %v\n", err)
		return
	}

	for _, fs := range fsResult.FileSystems {
		fsID := aws.ToString(fs.FileSystemId)
		currentTags := e.getCurrentTagsEFS(ctx, client, fsID)

		nameValue := currentTags["Name"]
		if nameValue == "" {
			nameValue = aws.ToString(fs.Name)
			if nameValue == "" {
				nameValue = fsID
			}
		}

		machineKey := normalizeKey(nameValue)
		if machineKey == "" {
			machineKey = fsID
		}

		tagsToAdd := []efstypes.Tag{}
		if _, exists := currentTags["Name"]; !exists {
			tagsToAdd = append(tagsToAdd, efstypes.Tag{Key: aws.String("Name"), Value: aws.String(nameValue)})
		}
		if _, exists := currentTags[machineKey]; !exists {
			tagsToAdd = append(tagsToAdd, efstypes.Tag{Key: aws.String(machineKey), Value: aws.String("")})
		}

		e.planOrApplyEFS(ctx, client, fsID, tagsToAdd, "EFS FileSystem")

		// Access Points
		apResult, err := client.DescribeAccessPoints(ctx, &efs.DescribeAccessPointsInput{
			FileSystemId: fs.FileSystemId,
		})
		if err == nil {
			for _, ap := range apResult.AccessPoints {
				apID := aws.ToString(ap.AccessPointId)
				apTags := e.getCurrentTagsEFS(ctx, client, apID)

				apName := apTags["Name"]
				if apName == "" {
					apName = fmt.Sprintf("%s-ap", nameValue)
				}

				apKey := normalizeKey(apName)
				if apKey == "" {
					apKey = apID
				}

				apTagsToAdd := []efstypes.Tag{}
				if _, exists := apTags["Name"]; !exists {
					apTagsToAdd = append(apTagsToAdd, efstypes.Tag{Key: aws.String("Name"), Value: aws.String(apName)})
				}
				if _, exists := apTags[apKey]; !exists {
					apTagsToAdd = append(apTagsToAdd, efstypes.Tag{Key: aws.String(apKey), Value: aws.String("")})
				}

				e.planOrApplyEFS(ctx, client, apID, apTagsToAdd, "EFS AccessPoint")
			}
		}
	}
}

// getCurrentTagsEFS gets current tags for an EFS resource
func (e *Engine) getCurrentTagsEFS(ctx context.Context, client *efs.Client, resourceID string) map[string]string {
	result, err := client.ListTagsForResource(ctx, &efs.ListTagsForResourceInput{
		ResourceId: aws.String(resourceID),
	})
	if err != nil {
		return make(map[string]string)
	}

	tags := make(map[string]string)
	for _, tag := range result.Tags {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return tags
}

// planOrApplyEFS applies tags to EFS resources
func (e *Engine) planOrApplyEFS(ctx context.Context, client *efs.Client, resourceID string, tags []efstypes.Tag, resourceType string) {
	if len(tags) == 0 {
		return
	}

	action := "PLAN"
	if e.opts.Apply {
		action = "APPLY"
	}

	shortID := resourceID
	if idx := strings.LastIndex(resourceID, "/"); idx >= 0 {
		shortID = resourceID[idx+1:]
	}

	for _, tag := range tags {
		value := aws.ToString(tag.Value)
		if value == "" {
			value = "(empty)"
		}
		fmt.Printf("    [%s] %s %s → %s = %s\n", action, resourceType, shortID, aws.ToString(tag.Key), value)
	}

	if !e.opts.Apply {
		return
	}

	_, err := client.TagResource(ctx, &efs.TagResourceInput{
		ResourceId: aws.String(resourceID),
		Tags:       tags,
	})
	if err != nil {
		fmt.Printf("    [ERROR] %s %s: %v\n", resourceType, shortID, err)
	}
}

// processFSx processes FSx resources in a region
func (e *Engine) processFSx(ctx context.Context, region string) {
	mode := "DRY-RUN"
	if e.opts.Apply {
		mode = "APPLY"
	}

	fmt.Printf("\n[FSx] Processing FSx resources in %s (%s)\n", strings.ToUpper(region), mode)

	regionCfg := e.cfg.Copy()
	regionCfg.Region = region
	client := fsx.NewFromConfig(regionCfg)

	// File Systems
	fsResult, err := client.DescribeFileSystems(ctx, &fsx.DescribeFileSystemsInput{})
	if err != nil {
		fmt.Printf("[WARN] Failed to describe FSx file systems: %v\n", err)
		return
	}

	for _, fs := range fsResult.FileSystems {
		fsARN := aws.ToString(fs.ResourceARN)
		currentTags := e.getCurrentTagsFSx(ctx, client, fsARN)

		nameValue := currentTags["Name"]
		if nameValue == "" {
			nameValue = aws.ToString(fs.FileSystemId)
		}

		machineKey := normalizeKey(nameValue)
		if machineKey == "" {
			machineKey = aws.ToString(fs.FileSystemId)
		}

		tagsToAdd := []fsxtypes.Tag{}
		if _, exists := currentTags["Name"]; !exists {
			tagsToAdd = append(tagsToAdd, fsxtypes.Tag{Key: aws.String("Name"), Value: aws.String(nameValue)})
		}
		if _, exists := currentTags[machineKey]; !exists {
			tagsToAdd = append(tagsToAdd, fsxtypes.Tag{Key: aws.String(machineKey), Value: aws.String("")})
		}

		e.planOrApplyFSx(ctx, client, fsARN, tagsToAdd, "FSx FileSystem")
	}

	// Backups
	backupResult, err := client.DescribeBackups(ctx, &fsx.DescribeBackupsInput{})
	if err == nil {
		for _, backup := range backupResult.Backups {
			backupARN := aws.ToString(backup.ResourceARN)
			currentTags := e.getCurrentTagsFSx(ctx, client, backupARN)

			nameValue := currentTags["Name"]
			if nameValue == "" {
				nameValue = aws.ToString(backup.BackupId)
			}

			machineKey := normalizeKey(nameValue)
			if machineKey == "" {
				machineKey = aws.ToString(backup.BackupId)
			}

			tagsToAdd := []fsxtypes.Tag{}
			if _, exists := currentTags["Name"]; !exists {
				tagsToAdd = append(tagsToAdd, fsxtypes.Tag{Key: aws.String("Name"), Value: aws.String(nameValue)})
			}
			if _, exists := currentTags[machineKey]; !exists {
				tagsToAdd = append(tagsToAdd, fsxtypes.Tag{Key: aws.String(machineKey), Value: aws.String("")})
			}

			e.planOrApplyFSx(ctx, client, backupARN, tagsToAdd, "FSx Backup")
		}
	}

	// Volumes
	volumeResult, err := client.DescribeVolumes(ctx, &fsx.DescribeVolumesInput{})
	if err == nil {
		for _, volume := range volumeResult.Volumes {
			volumeARN := aws.ToString(volume.ResourceARN)
			currentTags := e.getCurrentTagsFSx(ctx, client, volumeARN)

			nameValue := currentTags["Name"]
			if nameValue == "" {
				nameValue = aws.ToString(volume.VolumeId)
			}

			machineKey := normalizeKey(nameValue)
			if machineKey == "" {
				machineKey = aws.ToString(volume.VolumeId)
			}

			tagsToAdd := []fsxtypes.Tag{}
			if _, exists := currentTags["Name"]; !exists {
				tagsToAdd = append(tagsToAdd, fsxtypes.Tag{Key: aws.String("Name"), Value: aws.String(nameValue)})
			}
			if _, exists := currentTags[machineKey]; !exists {
				tagsToAdd = append(tagsToAdd, fsxtypes.Tag{Key: aws.String(machineKey), Value: aws.String("")})
			}

			e.planOrApplyFSx(ctx, client, volumeARN, tagsToAdd, "FSx Volume")
		}
	}
}

// getCurrentTagsFSx gets current tags for an FSx resource
func (e *Engine) getCurrentTagsFSx(ctx context.Context, client *fsx.Client, resourceARN string) map[string]string {
	result, err := client.ListTagsForResource(ctx, &fsx.ListTagsForResourceInput{
		ResourceARN: aws.String(resourceARN),
	})
	if err != nil {
		return make(map[string]string)
	}

	tags := make(map[string]string)
	for _, tag := range result.Tags {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return tags
}

// planOrApplyFSx applies tags to FSx resources
func (e *Engine) planOrApplyFSx(ctx context.Context, client *fsx.Client, resourceARN string, tags []fsxtypes.Tag, resourceType string) {
	if len(tags) == 0 {
		return
	}

	action := "PLAN"
	if e.opts.Apply {
		action = "APPLY"
	}

	shortID := resourceARN
	if idx := strings.LastIndex(resourceARN, "/"); idx >= 0 {
		shortID = resourceARN[idx+1:]
	}

	for _, tag := range tags {
		value := aws.ToString(tag.Value)
		if value == "" {
			value = "(empty)"
		}
		fmt.Printf("    [%s] %s %s → %s = %s\n", action, resourceType, shortID, aws.ToString(tag.Key), value)
	}

	if !e.opts.Apply {
		return
	}

	_, err := client.TagResource(ctx, &fsx.TagResourceInput{
		ResourceARN: aws.String(resourceARN),
		Tags:        tags,
	})
	if err != nil {
		fmt.Printf("    [ERROR] %s %s: %v\n", resourceType, shortID, err)
	}
}
