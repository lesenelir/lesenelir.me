//go:build darwin

package app

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/store"
)

// Kit began as Voice, a voice input app on its own. Its data lived in
// Voice's own directory, its keys under Voice's identifier in the
// Keychain.
const (
	voiceIdentifier = "me.lesenelir.voice"
	voiceSection    = "voice"
)

// voiceSecrets are the Keychain accounts Voice kept its API keys under.
var voiceSecrets = []string{"openrouter-api-key", "soniox-api-key"}

// migrateFromVoice brings what Voice kept into a new install of Kit: its
// settings, history and API keys. Voice's own data is left as it was.
func migrateFromVoice(dir string, st *store.Store) {
	old := voiceDataDir(dir)
	ok, err := migrateFiles(old, dir, st)
	if err != nil {
		log.Printf("migrate from Voice: %v", err)
	}
	if !ok {
		return
	}
	service := voiceIdentifier
	if strings.HasSuffix(filepath.Base(dir), " Dev") {
		service += ".dev"
	}
	migrateSecrets(mygo.Secrets.Service(service), mygo.Secrets)
	log.Printf("migrated from Voice: %s", old)
}

// migrateFiles copies Voice's settings in old into st, split into the
// app's and voice input's sections, and its history and what it must
// still restore into dir's voice directory. It reports whether there was
// a Voice to migrate from.
func migrateFiles(old, dir string, st *store.Store) (bool, error) {
	raw, err := os.ReadFile(filepath.Join(old, "settings.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !json.Valid(raw) {
		return false, errors.New("settings.json is not JSON")
	}
	// Voice kept its settings flat, the app's and voice input's together;
	// each section reads its own fields and ignores the others.
	for _, section := range []string{settingsSection, voiceSection} {
		if err := st.SetRaw(section, raw); err != nil {
			return true, err
		}
	}
	var errs []error
	for _, name := range []string{"history.jsonl", "output-to-restore.json"} {
		errs = append(errs, copyFile(filepath.Join(old, name), filepath.Join(dir, voiceSection, name)))
	}
	return true, errors.Join(errs...)
}

// migrateSecrets copies Voice's API keys from its Keychain service into
// Kit's, unless Kit has them already.
func migrateSecrets(from, to mygo.SecretsModule) {
	for _, name := range voiceSecrets {
		if v, err := to.Get(name); err == nil && len(v) > 0 {
			continue
		}
		v, err := from.Get(name)
		if err != nil || len(v) == 0 {
			continue
		}
		if err := to.Set(name, v); err != nil {
			log.Printf("migrate from Voice: keychain %s: %v", name, err)
		}
	}
}

// voiceDataDir is where Voice kept its data, beside Kit's: "Voice" for
// "Kit", "Voice Dev" for "Kit Dev".
func voiceDataDir(dir string) string {
	name := filepath.Base(dir)
	return filepath.Join(filepath.Dir(dir), strings.Replace(name, "Kit", "Voice", 1))
}

// copyFile copies src to dst, if src exists.
func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}
