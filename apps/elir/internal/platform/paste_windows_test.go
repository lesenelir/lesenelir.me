package platform

import (
	"errors"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
)

func TestPasteStopsBeforeInjectionWhenClipboardFails(t *testing.T) {
	read, write, press := readClipboard, writeClipboard, injectControl
	t.Cleanup(func() { readClipboard, writeClipboard, injectControl = read, write, press })
	failure := errors.New("clipboard locked")
	for _, failRead := range []bool{true, false} {
		writes, presses := 0, 0
		readClipboard = func(...transfer.Format) (transfer.Data, error) {
			if failRead {
				return transfer.Data{}, failure
			}
			return transfer.TextData("previous clipboard"), nil
		}
		writeClipboard = func(transfer.Data, ...mygo.ClipboardOptions) error { writes++; return failure }
		injectControl = func(uint16) error { presses++; return nil }
		if err := (Paster{}).Paste("result"); !errors.Is(err, failure) {
			t.Fatalf("error %v", err)
		}
		if presses != 0 || failRead && writes != 0 {
			t.Fatalf("writes=%d presses=%d after clipboard failure", writes, presses)
		}
	}
}

func TestTemporaryPasteOptsOutOfWindowsHistory(t *testing.T) {
	read, write, press := readClipboard, writeClipboard, injectControl
	t.Cleanup(func() { readClipboard, writeClipboard, injectControl = read, write, press })
	readClipboard = func(...transfer.Format) (transfer.Data, error) { return transfer.Data{}, nil }
	var written transfer.Data
	writeClipboard = func(data transfer.Data, _ ...mygo.ClipboardOptions) error { written = data; return nil }
	// Stop before scheduling restoration, without injecting real input.
	injectControl = func(uint16) error { return ErrNotTrusted }
	if err := (Paster{}).Paste("private transcript"); !errors.Is(err, ErrNotTrusted) {
		t.Fatal(err)
	}
	items := written.Items()
	if len(items) != 1 {
		t.Fatal("no clipboard item")
	}
	text, err := items[0].Read(transfer.Text)
	if err != nil || string(text) != "private transcript" {
		t.Fatal("missing paste text")
	}
	marker, err := items[0].Read("ExcludeClipboardContentFromMonitorProcessing")
	if err != nil || len(marker) == 0 {
		t.Fatal("temporary text can enter Windows clipboard history")
	}
}
