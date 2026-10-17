package app

import (
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/secret"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
)

// Earlier releases were macOS-only; Windows has no legacy data to migrate.
func migrateFromKit(string) (bool, error)                  { return false, nil }
func migrateFromVoice(string, *store.Store, *secret.Store) {}
