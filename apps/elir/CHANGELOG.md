# Changelog

Elir's versions, newest first; up to 0.2.0 the app was called Kit. `scripts/release.sh` takes the section of the version it releases as the release notes.

## 0.2.2

- The menu bar icon is now the head of the character of Elir's icon, tinted to match a light or dark menu bar.

## 0.2.1

- Kit is now **Elir**, with a new icon. Settings, history, screenshots, API keys and permissions carry over. An
  installed copy that updates itself keeps its file name, Kit.app, which can be renamed in Finder.
- Settings → General → Show in Dock gives Elir a Dock icon, and a place in ⌘Tab, besides its menu bar icon.

## 0.2.0

- Kit updates itself: Settings → General checks for a new version, shows what's new and installs it with one click,
  then relaunches. It checks by itself once a day, which the same page can turn off; the menu bar menu shows an update
  that is waiting.
- Coming from 0.1.0, download this version once by hand: 0.1.0 cannot update itself.

## 0.1.0

The first release of Kit, a menu bar toolbox for macOS.

### Voice input

- `Fn` dictates: speech is transcribed, cleaned up and pasted where the cursor is.
- `Fn` + `Left Shift` translates instead; `Tab` or a click on the voice bar switches the target language, `Esc`
  cancels.
- Soniox transcribes while you speak, with OpenRouter as the fallback; OpenRouter cleans up and translates.
- Music and videos are muted or lowered while recording.
- A dictionary of names and terms, and a searchable history of the text, which can be turned off.

### Screenshots

- `⌘⇧2` captures an area or, with `Space`, a window, as macOS's `⌘⇧4` does.
- The screenshot is copied to the clipboard (paste with `Control+V` in Claude Code and other terminals), and its
  thumbnail can be dragged into any app.
- Screenshots are kept in Kit, never on the Desktop, for 1, 7 or 30 days or forever, and listed in Settings.

### The app

- English and Simplified Chinese, light and dark, and an accent color.
- Open at login.
- Brings over the settings, history and API keys of Voice, the app Kit grew from.
