package auth

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

// LoadCredentials loads credentials and region with zero external deps.
// Order:
// 1) Environment variables
// 2) ~/.aws/credentials (profile "default" by default)
// 3) ~/.aws/config (region)
func LoadCredentials(profile string) (Credentials, error) {
	if profile == "" {
		profile = "default"
	}

	// 1) ENV first
	c := Credentials{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
		Region:          os.Getenv("AWS_REGION"),
	}
	if c.Region == "" {
		c.Region = os.Getenv("AWS_DEFAULT_REGION")
	}

	// If env has the keys, we can accept early (region can be filled later).
	if c.AccessKeyID != "" && c.SecretAccessKey != "" {
		if c.Region == "" {
			// Try fill region from config file
			region, _ := readAWSRegion(profile)
			c.Region = region
		}
		if c.Region == "" {
			c.Region = "us-east-1" // sensible default, adjust if you want strict mode
		}
		return c, nil
	}

	// 2) Shared credentials file
	fileCreds, err := readAWSCredentials(profile)
	if err == nil {
		if c.AccessKeyID == "" {
			c.AccessKeyID = fileCreds.AccessKeyID
		}
		if c.SecretAccessKey == "" {
			c.SecretAccessKey = fileCreds.SecretAccessKey
		}
		if c.SessionToken == "" {
			c.SessionToken = fileCreds.SessionToken
		}
	}

	// 3) Region from config file
	if c.Region == "" {
		region, _ := readAWSRegion(profile)
		c.Region = region
	}
	if c.Region == "" {
		c.Region = "us-east-1"
	}

	if c.AccessKeyID == "" || c.SecretAccessKey == "" {
		return Credentials{}, errors.New("missing AWS credentials (env or ~/.aws/credentials)")
	}

	return c, nil
}

func readAWSCredentials(profile string) (Credentials, error) {
	path := awsFilePath("credentials")
	return parseSimpleINI(path, profile, false)
}

func readAWSRegion(profile string) (string, error) {
	path := awsFilePath("config")
	c, err := parseSimpleINI(path, profile, true)
	if err != nil {
		return "", err
	}
	return c.Region, nil
}

func awsFilePath(name string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", name)
}

// parseSimpleINI is a tiny INI reader adequate for AWS shared files.
// It supports:
// - [default]
// - [profile myprofile] in config
// - keys: aws_access_key_id, aws_secret_access_key, aws_session_token, region
func parseSimpleINI(path, profile string, isConfig bool) (Credentials, error) {
	f, err := os.Open(path)
	if err != nil {
		return Credentials{}, err
	}
	defer f.Close()

	targetSection := profile
	if isConfig && profile != "default" {
		// AWS config uses [profile name] except default
		targetSection = "profile " + profile
	}

	var current string
	out := Credentials{}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}

		if current != targetSection {
			continue
		}

		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		switch key {
		case "aws_access_key_id":
			out.AccessKeyID = val
		case "aws_secret_access_key":
			out.SecretAccessKey = val
		case "aws_session_token":
			out.SessionToken = val
		case "region":
			out.Region = val
		}
	}

	if err := sc.Err(); err != nil {
		return Credentials{}, err
	}

	// It's ok if region-only parsing returns empty keys.
	return out, nil
}