package secret

import (
	"errors"
	"github.com/egoist/mygo"
	"log"
	"os"
	"strings"
)

func readSecretFile(path string) ([]byte, error)     { return os.ReadFile(path) }
func writeSecretFile(path string, data []byte) error { return os.WriteFile(path, data, 0600) }

// fromKeychain reads a key the app kept in the Keychain before, "" when
// there is none.
func fromKeychain(name string) string {
	b, err := mygo.Secrets.Get(name)
	if err != nil && !errors.Is(err, mygo.ErrSecretNotFound) {
		log.Printf("keychain %s: %v", name, err)
	}
	return strings.TrimSpace(string(b))
}
