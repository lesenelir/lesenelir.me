package widget

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/i18n"
)

// FormatTime writes when something was made in the language's way, the
// month and day before the time: 10月17日 14:32, Oct 17, 2:32 PM.
func FormatTime(locale string, t time.Time) string {
	return i18n.FormatDate(locale, t.Local(), i18n.DateOptions{Skeleton: "MMMdjm"})
}

// FormatSize writes a file size: 512 B, 12.4 KB, 1.3 MB.
func FormatSize(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
}
