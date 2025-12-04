package tagging

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

func TestResolveRegions_PrefersExplicitRegionsSlice(t *testing.T) {
	e := &Engine{
		opts: Options{
			Regions: []string{"us-east-1", "us-east-2"},
		},
	}

	regions := e.resolveRegions(context.Background())
	if len(regions) != 2 {
		t.Fatalf("expected 2 regions, got %d", len(regions))
	}
	if regions[0] != "us-east-1" || regions[1] != "us-east-2" {
		t.Errorf("unexpected regions: %#v", regions)
	}
}

func TestResolveRegions_UsesSingleRegionWhenProvided(t *testing.T) {
	e := &Engine{
		opts: Options{
			Region: "eu-west-3",
		},
	}

	regions := e.resolveRegions(context.Background())
	if len(regions) != 1 {
		t.Fatalf("expected 1 region, got %d", len(regions))
	}
	if regions[0] != "eu-west-3" {
		t.Errorf("expected region eu-west-3, got %s", regions[0])
	}
}

func TestResolveRegions_FallsBackToTargetRegions(t *testing.T) {
	// No explicit Regions and no Region → should fall back to global TargetRegions
	e := &Engine{
		opts: Options{},
	}

	regions := e.resolveRegions(context.Background())
	if len(regions) != len(TargetRegions) {
		t.Fatalf("expected %d regions from TargetRegions, got %d", len(TargetRegions), len(regions))
	}

	for i, r := range TargetRegions {
		if regions[i] != r {
			t.Errorf("expected regions[%d] = %s, got %s", i, r, regions[i])
		}
	}
}

func TestNormalizeKey_BasicCases(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"VA-WEB-ASG DEV TEST", "VA-WEB-ASG-DEV-TEST"},
		{"  Name!!  ", "Name--"},
		{"simple", "simple"},
		{"multi   space", "multi---space"},
		{"UPPER_case-123", "UPPER-case-123"},
	}

	for _, tc := range cases {
		got := normalizeKey(tc.in)
		if got != tc.want {
			t.Errorf("normalizeKey(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestGetNameTag(t *testing.T) {
	instance := ec2types.Instance{
		Tags: []ec2types.Tag{
			{Key: aws.String("Env"), Value: aws.String("test")},
			{Key: aws.String("Name"), Value: aws.String("VA-WEB-ASG-DEV-TEST")},
		},
	}

	e := &Engine{}
	got := e.getNameTag(instance)
	if got != "VA-WEB-ASG-DEV-TEST" {
		t.Fatalf("expected Name tag 'VA-WEB-ASG-DEV-TEST', got %q", got)
	}
}

func TestGetNameTag_EmptyWhenMissing(t *testing.T) {
	instance := ec2types.Instance{
		Tags: []ec2types.Tag{
			{Key: aws.String("Env"), Value: aws.String("test")},
		},
	}

	e := &Engine{}
	got := e.getNameTag(instance)
	if got != "" {
		t.Fatalf("expected empty Name when tag is missing, got %q", got)
	}
}

func TestGetMachineKey_UsesNormalizedName(t *testing.T) {
	instance := ec2types.Instance{
		InstanceId: aws.String("i-1234567890abcdef0"),
		Tags: []ec2types.Tag{
			{Key: aws.String("Name"), Value: aws.String("VA WEB ASG DEV TEST")},
		},
	}

	e := &Engine{}
	got := e.getMachineKey(instance)
	// "VA WEB ASG DEV TEST" → "VA-WEB-ASG-DEV-TEST"
	if got != "VA-WEB-ASG-DEV-TEST" {
		t.Fatalf("expected machine key 'VA-WEB-ASG-DEV-TEST', got %q", got)
	}
}

func TestGetMachineKey_FallsBackToInstanceID(t *testing.T) {
	instance := ec2types.Instance{
		InstanceId: aws.String("i-1234567890abcdef0"),
		Tags:       []ec2types.Tag{}, // no Name tag
	}

	e := &Engine{}
	got := e.getMachineKey(instance)
	if got != "i-1234567890abcdef0" {
		t.Fatalf("expected machine key to fall back to instance ID, got %q", got)
	}
}

func TestBuildCostAllocationTagStatus(t *testing.T) {
	keys := []string{"Name", "Machine", "Env"}

	entries := buildCostAllocationTagStatus(keys)
	if len(entries) != len(keys) {
		t.Fatalf("expected %d entries, got %d", len(keys), len(entries))
	}

	for i, key := range keys {
		if entries[i].TagKey == nil || *entries[i].TagKey != key {
			t.Errorf("entry[%d] TagKey = %v, want %q", i, entries[i].TagKey, key)
		}
		if entries[i].Status != ceTypes.CostAllocationTagStatusActive {
			t.Errorf("entry[%d] Status = %v, want %v", i, entries[i].Status, ceTypes.CostAllocationTagStatusActive)
		}
	}
}