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

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/secret"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
)

// Elir was Kit, which began as Voice, a voice input app on its own. Kit
// had Elir's identifier, so its Keychain keys and permissions are Elir's
// already, but its data lived in a directory named after it. Voice's data
// lived in its own directory too, its keys under Voice's identifier in the
// Keychain.
const (
	voiceIdentifier = "me.lesenelir.voice"
	voiceSection    = "voice"
)

// voiceSecrets are the Keychain accounts Voice kept its API keys under.
var voiceSecrets = []string{"openrouter-api-key", "soniox-api-key"}

// migrateFromKit moves Kit's data into Elir's directory: its settings,
// history, screenshots and what it must still restore. It moves rather
// than copies, as screenshots take room, what Elir has not made already,
// and reports whether there was a Kit to migrate from. It runs before the
// settings are opened.
func migrateFromKit(dir string) (bool, error) {
	old := previousDataDir(dir, "Kit")
	if fileExists(filepath.Join(dir, "settings.json")) || !fileExists(filepath.Join(old, "settings.json")) {
		return false, nil
	}
	entries, err := os.ReadDir(old)
	if err != nil {
		return false, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, err
	}
	var errs []error
	for _, e := range entries {
		dst := filepath.Join(dir, e.Name())
		if _, err := os.Lstat(dst); err == nil {
			continue
		}
		errs = append(errs, os.Rename(filepath.Join(old, e.Name()), dst))
	}
	return true, errors.Join(errs...)
}

// migrateFromVoice brings what Voice kept into a new install of Elir: its
// settings, history and API keys. Voice's own data is left as it was.
func migrateFromVoice(dir string, st *store.Store, secrets *secret.Store) {
	old := previousDataDir(dir, "Voice")
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
	migrateSecrets(mygo.Secrets.Service(service), secrets)
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
// Elir's file of keys, unless Elir has them already.
func migrateSecrets(from mygo.SecretsModule, to *secret.Store) {
	for _, name := range voiceSecrets {
		v, err := from.Get(name)
		if err != nil || len(v) == 0 {
			continue
		}
		if _, err := to.Import(name, strings.TrimSpace(string(v))); err != nil {
			log.Printf("migrate from Voice: %s: %v", name, err)
		}
	}
}

// previousDataDir is where the app kept its data under an earlier name,
// beside dir: "Kit" for "Elir", "Voice Dev" for "Elir Dev".
func previousDataDir(dir, name string) string {
	if strings.HasSuffix(filepath.Base(dir), " Dev") {
		name += " Dev"
	}
	return filepath.Join(filepath.Dir(dir), name)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
