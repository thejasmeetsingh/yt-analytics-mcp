package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/youtubeanalytics/v2"
)

const (
	tokenFileName = "token.json"
	tokenFileMode = 0600 // Read/write for owner only
)

var (
	// Escape Codes
	Reset = "\033[0m"
	Red   = "\033[31m"
	Green = "\033[32m"
	Blue  = "\033[34m"
	Cyan  = "\033[36m"

	ErrReadingCredentials = fmt.Errorf("failed to read credentials file")
	ErrParsingCredentials = fmt.Errorf("failed to parse credentials")
	ErrReadingToken       = fmt.Errorf("failed to read token file")
	ErrGeneratingToken    = fmt.Errorf("failed to generate token from web")
	ErrSavingToken        = fmt.Errorf("failed to save token")
	ErrReadingAuthCode    = fmt.Errorf("failed to read authorization code")
	ErrExchangingAuthCode = fmt.Errorf("failed to exchange authorization code for token")
)

func GetConfig(path string) (*oauth2.Config, error) {
	credentialsData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w at path '%s': %v", ErrReadingCredentials, path, err)
	}

	// Configure OAuth2 with YouTube API scopes
	config, err := google.ConfigFromJSON(
		credentialsData,
		youtube.YoutubeReadonlyScope,              // Read-only access to YouTube account
		youtube.YoutubepartnerScope,               // Partner-level access
		youtube.YoutubeForceSslScope,              // Force SSL access
		youtubeanalytics.YtAnalyticsReadonlyScope, // Analytics data access
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParsingCredentials, err)
	}

	return config, nil
}

func GetClient(rootPath string, config *oauth2.Config, generateToken bool) (*http.Client, error) {
	tokenPath := filepath.Join(rootPath, tokenFileName)

	// Try to load existing token
	token, err := loadTokenFromFile(tokenPath)

	if err != nil || generateToken {
		// Generate new token via web OAuth flow
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}

		// Save the newly generated token
		if err := saveToken(tokenPath, token); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}

	// Create authenticated HTTP client
	return config.Client(context.Background(), token), nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println(Red + "\n=== YouTube API Authorization Required ===\n" + Reset)
	fmt.Println(Green + "Please visit this URL in your browser:" + Reset)
	fmt.Println(Blue + authURL + Reset + "\n")
	fmt.Print(Cyan + "After authorizing, paste the authorization code here: " + Reset)

	// Read authorization code from user input
	var authCode string
	if _, err := fmt.Scanln(&authCode); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadingAuthCode, err)
	}

	// Exchange authorization code for access token
	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchangingAuthCode, err)
	}

	fmt.Println("✓ Authorization successful!")
	return token, nil
}

func loadTokenFromFile(filePath string) (*oauth2.Token, error) {
	file, err := os.Open(filePath)

	// Create token.json file if it doesn't exists
	if os.IsNotExist(err) {
		if err := os.WriteFile(filePath, []byte(""), tokenFileMode); err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, fmt.Errorf("cannot open token file '%s': %v", filePath, err)
	}
	defer file.Close()

	token := &oauth2.Token{}
	if err := json.NewDecoder(file).Decode(token); err != nil {
		return nil, fmt.Errorf("cannot decode token JSON from '%s': %v", filePath, err)
	}

	return token, nil
}

func saveToken(filePath string, token *oauth2.Token) error {
	fmt.Printf("Saving credentials to: %s\n", filePath)

	// Create file with secure permissions (owner read/write only)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, tokenFileMode)
	if err != nil {
		return fmt.Errorf("%w to '%s': %v", ErrSavingToken, filePath, err)
	}
	defer file.Close()

	// Serialize token to JSON
	if err := json.NewEncoder(file).Encode(token); err != nil {
		return fmt.Errorf("%w - JSON encoding failed: %v", ErrSavingToken, err)
	}

	return nil
}
