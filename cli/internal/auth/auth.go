package auth

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const keyChainServiceName = "wisp-cli"

// Retrieves a GitHub token for the specified hostname.
// For "github.com" this would be the PAT you registered with wisp for github.com
func GetTokenForHost(host string) string {
	// Retrieve the environment variable token.
	if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		return envToken
	}

	// Retrieve the token from the os keyring.
	keyChainToken, err := keyring.Get(keyChainServiceName, host)
	if err == nil && keyChainToken != "" {
		return keyChainToken
	}

	// No token was found.
	return ""
}

// Stores a token for a specific host to the os keyring.
func StoreToken(host string, token string) error {
	err := keyring.Set(keyChainServiceName, host, token)
	if err != nil {
		return fmt.Errorf("failed to save token to keychain: %v", err)
	}
	return nil
}
