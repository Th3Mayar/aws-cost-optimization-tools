package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

var (
	currentProfile string
	currentConfig  aws.Config
)

// Login performs interactive AWS credential configuration
func Login(profile string) error {
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	if profile == "" {
		profile = "default"
	}

	fmt.Println(cyan("🔐 AWS Credentials Setup"))
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
	fmt.Println(green("✓ Credentials saved successfully!"))
	fmt.Println(yellow(fmt.Sprintf("Profile: %s", profile)))
	fmt.Println(yellow(fmt.Sprintf("Region: %s", region)))
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

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Println(green("✓ Successfully authenticated!"))
	fmt.Println(cyan(fmt.Sprintf("Account: %s", *identity.Account)))
	fmt.Println(cyan(fmt.Sprintf("User: %s", *identity.Arn)))
	fmt.Println(cyan(fmt.Sprintf("Profile: %s", profile)))
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

	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println(blue("Current AWS Identity:"))
	fmt.Println(cyan("  Account:"), yellow(*identity.Account))
	fmt.Println(cyan("  User ARN:"), yellow(*identity.Arn))
	fmt.Println(cyan("  User ID:"), yellow(*identity.UserId))
	fmt.Println(cyan("  Profile:"), yellow(currentProfile))
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
	
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Println(green("✓ Logged out successfully"))
	fmt.Println()
}
