package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

// RandomCharacters is used as a source for NewRandomString.
type RandomCharacters string

const (
	// Alpha is lower-case and upper-case ASCII characters, as well as digits 0-9.
	Alpha RandomCharacters = "abcdefghijklmnopqrstuvqxyzABCDEFGHIJKLMONPQRSTUVWXYZ1234567890"
	// Digits is digits 0-9.
	Digits RandomCharacters = "0123456789"
)

// NewRandomString creates a new cryptographically safe random string from a source of characters.
func NewRandomString(letters RandomCharacters, length int) string {
	maxInt := big.NewInt(int64(len(letters)))
	s := make([]byte, length)
	for i := range s {
		n, err := rand.Int(rand.Reader, maxInt)
		if err != nil {
			// this should never happen, just pick something pseudorandom as fallback
			s[i] = letters[(i*3+37)%len(letters)]
		}
		s[i] = letters[n.Int64()]
	}
	return string(s)
}

// IsValidToken checks tokensFile to see if token is valid.
// A token is valid when it appears in the tokensFile.
// A token may optionally have a TTL, which is appended at the end of the token.
// For example, the tokens file may look like this:
//
//	token1
//	token2|35616531876
//
// In the file above, token1 is a valid token. token2 is valid until the unix timestamp 35616531876.
func IsValidToken(token string, tokensFile string) (isValidToken, hasTTL bool) {
	if token == "" {
		return false, false
	}
	token = strings.TrimSpace(token)
	if b, err := os.ReadFile(tokensFile); err == nil {
		knownTokens := strings.Split(string(b), "\n")
		for _, knownToken := range knownTokens {
			parts := strings.SplitN(strings.TrimSpace(knownToken), "|", 2)
			if parts[0] != token {
				continue
			}
			if len(parts) == 1 {
				return true, false
			}
			// token with expiry
			if len(parts) == 2 {
				timestamp, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					return false, true
				}
				return time.Now().Before(time.Unix(timestamp, 0)), true
			}
		}
	}
	return false, false
}

// AppendToken appends a token to a file.
// Token files contain a single token in each line.
func AppendToken(token string, tokensFile string, chownGroup string) error {
	f, err := os.OpenFile(tokensFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", tokensFile, err)
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("%s\n", token)); err != nil {
		return fmt.Errorf("failed to append token to %s: %w", tokensFile, err)
	}
	// TODO: consider whether permissions should be 0600 instead
	SetupPermissions(tokensFile, chownGroup)
	return nil
}

// RemoveToken removes a token from a tokens file, if it exists.
// RemoveToken will not return an error if the token does not exist.
// RemoveToken will only the first occurence of the token, if it exists multiple times in the tokens file.
// RemoveToken will return an error if it fails to read or write the tokens file.
func RemoveToken(token string, tokensFile string, chownGroup string) error {
	b, err := os.ReadFile(tokensFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", tokensFile, err)
	}
	existingTokens := strings.Split(string(b), "\n")
	if len(existingTokens) == 0 {
		return nil
	}
	for idx, tokenInFile := range existingTokens {
		if strings.HasPrefix(tokenInFile, token) {
			newTokens := append(existingTokens[:idx], existingTokens[idx+1:]...)
			if err = os.WriteFile(tokensFile, []byte(strings.Join(newTokens, "\n")), 0660); err != nil {
				return fmt.Errorf("failed to write %s: %w", tokensFile, err)
			}
			// TODO: consider whether permissions should be 0600 instead
			SetupPermissions(tokensFile, chownGroup)
			break
		}
	}
	return nil
}
