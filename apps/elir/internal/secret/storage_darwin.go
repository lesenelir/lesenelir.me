package secret

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/egoist/mygo"
)

// decodeSecretFile returns the file's JSON: on macOS the file is the JSON,
// kept readable by the user alone.
func decodeSecretFile(data []byte) ([]byte, error) { return data, nil }

func writeSecretFile(path string, data []byte) error { return os.WriteFile(path, data, 0o600) }

// fromKeychain reads a key the app kept in the Keychain before, "" when
// there is none.
func fromKeychain(name string) string {
	b, err := mygo.Secrets.Get(name)
	if err != nil && !errors.Is(err, mygo.ErrSecretNotFound) {
		log.Printf("keychain %s: %v", name, err)
	}
	return strings.TrimSpace(string(b))
}
