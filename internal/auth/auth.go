package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"gopkg.in/ini.v1"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

var (
	currentProfile string
	currentConfig  aws.Config
)

// Login performs interactive AWS credential configuration
func Login(profile string) error {
	if profile == "" {
		profile = "default"
	}

	fmt.Printf("%s%s🔐 AWS Credentials Setup%s\n", colorBold, colorCyan, colorReset)
	fmt.Println()

	// Get credentials from user
	var accessKey, secretKey, region string

	fmt.Print("AWS Access Key ID: ")
	fmt.Scanln(&accessKey)

	fmt.Print("AWS Secret Access Key: ")
	fmt.Scanln(&secretKey)

	fmt.Print("Default region [us-east-1]: ")
	fmt.Scanln(&region)
	if region == "" {
		region = "us-east-1"
	}

	// Get AWS config directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	awsDir := filepath.Join(home, ".aws")
	if err := os.MkdirAll(awsDir, 0700); err != nil {
		return fmt.Errorf("failed to create .aws directory: %w", err)
	}

	// Update credentials file
	credentialsPath := filepath.Join(awsDir, "credentials")
	credentials, err := ini.Load(credentialsPath)
	if err != nil {
		credentials = ini.Empty()
	}

	credSection, err := credentials.NewSection(profile)
	if err != nil {
		credSection = credentials.Section(profile)
	}
	credSection.Key("aws_access_key_id").SetValue(accessKey)
	credSection.Key("aws_secret_access_key").SetValue(secretKey)

	if err := credentials.SaveTo(credentialsPath); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	// Update config file
	configPath := filepath.Join(awsDir, "config")
	awsConfig, err := ini.Load(configPath)
	if err != nil {
		awsConfig = ini.Empty()
	}

	profileName := profile
	if profile != "default" {
		profileName = "profile " + profile
	}

	configSection, err := awsConfig.NewSection(profileName)
	if err != nil {
		configSection = awsConfig.Section(profileName)
	}
	configSection.Key("region").SetValue(region)
	configSection.Key("output").SetValue("json")

	if err := awsConfig.SaveTo(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("%s%s✓ Credentials saved successfully!%s\n", colorBold, colorGreen, colorReset)
	fmt.Printf("%sProfile: %s%s\n", colorYellow, profile, colorReset)
	fmt.Printf("%sRegion: %s%s\n", colorYellow, region, colorReset)
	fmt.Println()

	// Validate credentials
	if err := UseProfile(profile); err != nil {
		return fmt.Errorf("failed to validate credentials: %w", err)
	}

	return nil
}

// UseProfile switches to a different AWS profile
func UseProfile(profile string) error {
	ctx := context.Background()

	// Load AWS config with specified profile
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config for profile '%s': %w", profile, err)
	}

	// Validate credentials by calling STS GetCallerIdentity
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to validate credentials: %w", err)
	}

	currentProfile = profile
	currentConfig = cfg

	fmt.Printf("%s%s✓ Successfully authenticated!%s\n", colorBold, colorGreen, colorReset)
	fmt.Printf("%sAccount: %s%s\n", colorCyan, *identity.Account, colorReset)
	fmt.Printf("%sUser: %s%s\n", colorCyan, *identity.Arn, colorReset)
	fmt.Printf("%sProfile: %s%s\n", colorCyan, profile, colorReset)
	fmt.Println()

	return nil
}

// GetCurrentProfile returns the currently active profile
func GetCurrentProfile() string {
	if currentProfile == "" {
		return "default"
	}
	return currentProfile
}

// GetCurrentConfig returns the current AWS config
func GetCurrentConfig() (aws.Config, error) {
	if currentProfile == "" {
		return aws.Config{}, fmt.Errorf("no profile loaded. Please run 'login' first")
	}
	return currentConfig, nil
}

// Whoami displays current AWS identity information
func Whoami() error {
	if currentProfile == "" {
		fmt.Println("Not logged in. Run 'login' to authenticate.")
		return nil
	}

	ctx := context.Background()
	stsClient := sts.NewFromConfig(currentConfig)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to get identity: %w", err)
	}

	fmt.Printf("%s%sCurrent AWS Identity:%s\n", colorBold, colorBlue, colorReset)
	fmt.Printf("%s  Account:%s %s%s%s\n", colorCyan, colorReset, colorYellow, *identity.Account, colorReset)
	fmt.Printf("%s  User ARN:%s %s%s%s\n", colorCyan, colorReset, colorYellow, *identity.Arn, colorReset)
	fmt.Printf("%s  User ID:%s %s%s%s\n", colorCyan, colorReset, colorYellow, *identity.UserId, colorReset)
	fmt.Printf("%s  Profile:%s %s%s%s\n", colorCyan, colorReset, colorYellow, currentProfile, colorReset)
	fmt.Println()

	return nil
}

// ListProfiles lists all available AWS profiles
func ListProfiles() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	credentialsPath := filepath.Join(home, ".aws", "credentials")
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		return []string{}, nil
	}

	credentials, err := ini.Load(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	profiles := []string{}
	for _, section := range credentials.Sections() {
		if section.Name() != "DEFAULT" {
			profiles = append(profiles, section.Name())
		}
	}

	return profiles, nil
}

// Logout clears the current session
func Logout() {
	currentProfile = ""
	currentConfig = aws.Config{}

	fmt.Printf("%s%s✓ Logged out successfully%s\n", colorBold, colorGreen, colorReset)
	fmt.Println()
}