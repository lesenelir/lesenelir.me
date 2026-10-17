package widget

import "testing"

func TestFormatSize(t *testing.T) {
	for n, want := range map[int64]string{0: "0 B", 900: "900 B", 12700: "12.4 KB", 3 << 20: "3.0 MB"} {
		if got := FormatSize(n); got != want {
			t.Errorf("FormatSize(%d) = %q, want %q", n, got, want)
		}
	}
}
